import json

def gen2(idx, name, sql, params):
    if idx == len(params):
        fp = open(name+'.sql', 'w')
        fp.write(sql)
        fp.close()
        return
    for i in range(len(params[idx]['enums'])):
        e = params[idx]['enums'][i]
        if e not in params[idx]['template']:
            continue
        t = params[idx]['template'][e]
        gen2(idx+1, name+"-"+e, sql.replace(params[idx]['replaces'], t), params)

def gen(name, template_sql, params_json):
    params = json.loads(params_json)
    params = params['params']
    gen2(0, name, template_sql, params)

name = "trending-repos"
template_sql = '''
WITH stars AS (
  SELECT
      ge.repo_id AS repo_id,
      COUNT(1) AS total,
      COUNT(DISTINCT actor_id) AS actors,
      -- Calculate the score of each star according to the time of the star, the closer to the 
      -- current time, the higher the score got, the score range is between 2-5. Then sum the
      -- scores of all stars to get the total score obtained from the stars for the repository.
      SUM(
          GREATEST (
              LEAST (
                  (
                      (
                          TIMESTAMPDIFF(SECOND, DATE_SUB(NOW(), INTERVAL 1 MONTH), ge.created_at) / 
                          TIMESTAMPDIFF(SECOND, DATE_SUB(NOW(), INTERVAL 1 MONTH), NOW())
                      ) * (5 - 2)
                  ), 5
              ), 2
          )
      ) AS score
  FROM github_events ge
  WHERE
      -- Notice: In the GitHub events, WatchEvent means star, not watch.
      type = 'WatchEvent'
      AND (ge.created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH) AND ge.created_at <= NOW())
  GROUP BY ge.repo_id
  -- Exclude code repositories that use the same user to duplicate stars.
  HAVING actors > 0.9 * total
), forks AS (
  SELECT
      ge.repo_id AS repo_id,
      COUNT(1) AS total,
      COUNT(DISTINCT actor_id) AS actors,
      -- Calculate the score of each fork according to the time of the fork, the closer to the 
      -- current time, the higher the score got, the score range is between 1-4. Then sum the
      -- scores of all forks to get the total score obtained from the forks for the repository.
      SUM(
          GREATEST (
              LEAST (
                  (
                      (
                          TIMESTAMPDIFF(SECOND, DATE_SUB(NOW(), INTERVAL 1 MONTH), ge.created_at) / 
                          TIMESTAMPDIFF(SECOND, DATE_SUB(NOW(), INTERVAL 1 MONTH), NOW())
                      ) * (4 - 1)
                  ), 4
              ), 1
          )
      ) AS score
  FROM github_events ge
  WHERE
      type = 'ForkEvent'
      AND (ge.created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH) AND ge.created_at <= NOW())
  GROUP BY ge.repo_id
  -- Exclude code repositories that use the same user to duplicate froks.
  HAVING actors > 0.9 * total
), topRepos AS (
    SELECT
        r.repo_id,
        r.repo_name,
        r.primary_language,
        r.description,
        s.total AS stars_inc,
        IFNULL(f.total, 0) AS forks_inc,
        -- Calculate the composite score for the repository.
        SUM(
            s.score + 
            IFNULL(f.score, 0) +
            -- Give the new repository a higher score base.
            ABS(1 /  (1 + TIMESTAMPDIFF(YEAR, r.created_at, NOW()))) * 200
        ) AS total_score
    FROM github_repos r
        JOIN stars s ON r.repo_id = s.repo_id
        LEFT JOIN forks f ON r.repo_id = f.repo_id
    WHERE
        -- Filter rule: The repository must have at least 5 stars.
        stars > 5
        AND stars < 50000
        -- Filter rule: The repository must have at least 5 forks.
        AND forks > 5
        -- Filter rule: The repository must have pushed new code within the last three months.
        AND pushed_at > DATE_SUB(NOW(), INTERVAL 3 MONTH)
        -- Filter rule: Exclude some malicious new repositories.
        AND created_at < DATE_SUB(NOW(), INTERVAL 1 DAY)
        -- Filter rule: There should be no uncivilized words in the name of the repository.
        AND LOWER(repo_name) NOT LIKE '%fuck%'
        -- Filter by repository language.
        AND primary_language = 'Java'
        AND repo_name NOT IN (SELECT name FROM blacklist_repos)
        AND is_deleted = 0
    GROUP BY r.repo_id
    ORDER BY total_score DESC
    LIMIT 100
), pull_requests AS (
  SELECT
      ge.repo_id AS repo_id,
      COUNT(1) AS total
  FROM github_events ge
  JOIN topRepos tr ON ge.repo_id = tr.repo_id
  WHERE
      type = 'PullRequestEvent'
      AND action = 'opened'
      AND (ge.created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH) AND ge.created_at <= NOW())
      AND actor_login NOT LIKE '%[bot]'
  GROUP BY ge.repo_id
), pushes AS (
  SELECT
      ge.repo_id AS repo_id,
      COUNT(1) AS total
  FROM github_events ge
  JOIN topRepos tr ON ge.repo_id = tr.repo_id
  WHERE
      type = 'PushEvent'
      AND (ge.created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH) AND ge.created_at <= NOW())
  	  AND actor_login NOT LIKE '%[bot]'
  GROUP BY ge.repo_id
), repo_with_top_contributors AS (
    SELECT
        repo_id, SUBSTRING_INDEX(GROUP_CONCAT(DISTINCT actor_login ORDER BY cnt DESC SEPARATOR ','), ',', 5) AS actor_logins
    FROM (
        SELECT
            ge.repo_id AS repo_id,
            ge.actor_login AS actor_login,
            COUNT(*) AS cnt
        FROM github_events ge
        WHERE
            (
                (type = 'PullRequestEvent' AND action = 'opened') OR
                (type = 'IssuesEvent' AND action = 'opened') OR
                (type = 'PullRequestReviewEvent' AND action = 'created') OR
                (type = 'PushEvent' AND action = '')
            )
            AND (ge.created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH) AND ge.created_at <= NOW())
            AND ge.repo_id IN (SELECT tr.repo_id FROM topRepos tr)
            AND ge.actor_login NOT IN (SELECT bu.login FROM blacklist_users bu)
            AND ge.actor_login NOT LIKE '%bot%'
        GROUP BY ge.repo_id, ge.actor_login
        ORDER BY ge.repo_id, cnt DESC
    ) sub
    GROUP BY repo_id
), repo_with_collections AS (
    SELECT
        tr.repo_id, GROUP_CONCAT(DISTINCT c.name) AS collection_names
    FROM topRepos tr
    JOIN collection_items ci ON ci.repo_name = tr.repo_name
    JOIN collections c ON ci.collection_id = c.id
    WHERE c.public = true
    GROUP BY tr.repo_id
)
SELECT
    tr.repo_id,
    tr.repo_name,
    tr.primary_language AS language,
    tr.description,
    tr.stars_inc AS stars,
    tr.forks_inc AS forks,
    pr.total AS pull_requests,
    pu.total AS pushes,
    tr.total_score,
    tc.actor_logins AS contributor_logins,
    rc.collection_names AS collection_names
FROM
	topRepos tr
    LEFT JOIN repo_with_top_contributors tc ON tr.repo_id = tc.repo_id
    LEFT JOIN repo_with_collections rc ON tr.repo_id = rc.repo_id
    LEFT JOIN pull_requests pr ON tr.repo_id = pr.repo_id
    LEFT JOIN pushes pu ON tr.repo_id = pu.repo_id
ORDER BY total_score DESC
'''
params_json = '''
{
  "public": true,
  "description": "Repository trending alternative to https://github.com/trending",
  "cacheHours": -1,
  "refreshCron": {
    "param": "period",
    "on": {
      "past_24_hours": "@daily",
      "past_week": "@daily",
      "past_month": "@daily",
      "past_3_months": "@daily"
    }
  },
  "onlyFromCache": true,
  "params": [
    {
      "name": "period",
      "description": "Trending time period.",
      "replaces": "DATE_SUB(NOW(), INTERVAL 1 MONTH)",
      "default": "past_24_hours",
      "enums": [
       "past_24_hours", "past_week", "past_month", "past_3_months"
      ],
      "template": {
        "past_24_hours": "DATE_SUB(NOW(), INTERVAL 1 DAY)",
        "past_week": "DATE_SUB(NOW(), INTERVAL 1 WEEK)",
        "past_month": "DATE_SUB(NOW(), INTERVAL 1 MONTH)",
        "past_3_months": "DATE_SUB(NOW(), INTERVAL 3 MONTH)"
      }
    },
    {
      "name": "language",
      "description": "Primary language",
      "replaces": "AND primary_language = 'Java'",
      "default": "All",
      "enums": [
        "All", "JavaScript", "Java", "Python", "PHP", "C++", "C#", "Typescript", "Shell", "C", "Ruby", "Rust", "Go", "Kotlin", "HCL", "PowerShell", "CMake", "Groovy", "PLpgSQL", "TSQL", "Dart", "Swift", "HTML", "CSS", "Elixir", "Haskell", "Solidity", "Assembly", "R", "Scala", "Julia", "Lua", "Clojure", "Erlang", "Common Lisp", "Emacs Lisp", "OCaml", "MATLAB", "Objective-C", "Perl", "Fortran"
      ],
      "template": {
        "All": "",
        "JavaScript": "AND primary_language = 'JavaScript'",
        "Java": "AND primary_language = 'Java'",
        "Python": "AND primary_language = 'Python'",
        "PHP": "AND primary_language = 'PHP'",
        "C++": "AND primary_language = 'C++'",
        "C#": "AND primary_language = 'C#'",
        "TypeScript": "AND primary_language = 'TypeScript'",
        "Shell": "AND primary_language = 'Shell'",
        "C": "AND primary_language = 'C'",
        "Ruby": "AND primary_language = 'Ruby'",
        "Rust": "AND primary_language = 'Rust'",
        "Go": "AND primary_language = 'Go'",
        "Kotlin": "AND primary_language = 'Kotlin'",
        "HCL": "AND primary_language = 'HCL'",
        "PowerShell": "AND primary_language = 'PowerShell'",
        "CMake": "AND primary_language = 'CMake'",
        "Groovy": "AND primary_language = 'Groovy'",
        "PLpgSQL": "AND primary_language = 'PLpgSQL'",
        "TSQL": "AND primary_language = 'TSQL'",
        "Dart": "AND primary_language = 'Dart'",
        "Swift": "AND primary_language = 'Swift'",
        "HTML": "AND primary_language = 'HTML'",
        "CSS": "AND primary_language = 'CSS'",
        "Elixir": "AND primary_language = 'Elixir'",
        "Haskell": "AND primary_language = 'Haskell'",
        "Solidity": "AND primary_language = 'Solidity'",
        "Assembly": "AND primary_language = 'Assembly'",
        "R": "AND primary_language = 'R'",
        "Scala": "AND primary_language = 'Scala'",
        "Julia": "AND primary_language = 'Julia'",
        "Lua": "AND primary_language = 'Lua'",
        "Clojure": "AND primary_language = 'Clojure'",
        "Erlang": "AND primary_language = 'Erlang'",
        "Common Lisp": "AND primary_language = 'Common Lisp'",
        "Emacs Lisp": "AND primary_language = 'Emacs Lisp'",
        "OCaml": "AND primary_language = 'OCaml'",
        "MATLAB": "AND primary_language = 'MATLAB'",
        "Objective-C": "AND primary_language = 'Objective-C'",
        "Perl": "AND primary_language = 'Perl'",
        "Fortran": "AND primary_language = 'Fortran'"
      }
    }
  ],
  "resultSchema": {
    "type": "object",
    "properties": {
      "repo_id": {
        "type": "number"
      },
      "repo_name": {
        "type": "string"
      },
      "language": {
        "type": "string"
      },
      "description": {
        "type": "string"
      },
      "stars": {
        "type": "number"
      },
      "forks": {
        "type": "number"
      },
      "pull_requests": {
        "type": "number"
      },
      "pushes": {
        "type": "number"
      },
      "total_score": {
        "type": "number"
      },
      "contributor_logins": {
        "type": "string",
        "nullable": true
      },
      "collection_names": {
        "type": "string",
        "nullable": true
      }
    },
    "required": ["repo_id", "repo_name", "language", "description", "stars", "forks", "pull_requests", "pushes", "total_score", "contributor_logins", "collection_names"]
  }
}'''

gen(name, template_sql, params_json)