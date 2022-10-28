CREATE
DATABASE IF NOT EXISTS gharchive_dev;

USE
gharchive_dev;

CREATE TABLE `github_events`
(
    `id`                 bigint(20) DEFAULT NULL,
    `type`               varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `created_at`         datetime                                DEFAULT NULL,
    `repo_id`            bigint(20) DEFAULT NULL,
    `repo_name`          varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `actor_id`           bigint(20) DEFAULT NULL,
    `actor_login`        varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `actor_location`     varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `language`           varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `additions`          bigint(20) DEFAULT NULL,
    `deletions`          bigint(20) DEFAULT NULL,
    `action`             varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `number`             int(11) DEFAULT NULL,
    `commit_id`          varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `comment_id`         bigint(20) DEFAULT NULL,
    `org_login`          varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `org_id`             bigint(20) DEFAULT NULL,
    `state`              varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `closed_at`          datetime                                DEFAULT NULL,
    `comments`           int(11) DEFAULT NULL,
    `pr_merged_at`       datetime                                DEFAULT NULL,
    `pr_merged`          tinyint(1) DEFAULT NULL,
    `pr_changed_files`   int(11) DEFAULT NULL,
    `pr_review_comments` int(11) DEFAULT NULL,
    `pr_or_issue_id`     bigint(20) DEFAULT NULL,
    `event_day`          date                                    DEFAULT NULL,
    `event_month`        date                                    DEFAULT NULL,
    `author_association` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `event_year`         int(11) DEFAULT NULL,
    `push_size`          int(11) DEFAULT NULL,
    `push_distinct_size` int(11) DEFAULT NULL,
    KEY                  `index_github_events_on_id` (`id`),
    KEY                  `index_github_events_on_action` (`action`),
    KEY                  `index_github_events_on_actor_id` (`actor_id`),
    KEY                  `index_github_events_on_actor_login` (`actor_login`),
    KEY                  `index_github_events_on_additions` (`additions`),
    KEY                  `index_github_events_on_closed_at` (`closed_at`),
    KEY                  `index_github_events_on_comment_id` (`comment_id`),
    KEY                  `index_github_events_on_comments` (`comments`),
    KEY                  `index_github_events_on_commit_id` (`commit_id`),
    KEY                  `index_github_events_on_created_at` (`created_at`),
    KEY                  `index_github_events_on_deletions` (`deletions`),
    KEY                  `index_github_events_on_event_day` (`event_day`),
    KEY                  `index_github_events_on_event_month` (`event_month`),
    KEY                  `index_github_events_on_event_year` (`event_year`),
    KEY                  `index_github_events_on_language` (`language`),
    KEY                  `index_github_events_on_org_id` (`org_id`),
    KEY                  `index_github_events_on_org_login` (`org_login`),
    KEY                  `index_github_events_on_pr_changed_files` (`pr_changed_files`),
    KEY                  `index_github_events_on_pr_merged_at` (`pr_merged_at`),
    KEY                  `index_github_events_on_pr_or_issue_id` (`pr_or_issue_id`),
    KEY                  `index_github_events_on_pr_review_comments` (`pr_review_comments`),
    KEY                  `index_github_events_on_repo_id` (`repo_id`),
    KEY                  `index_github_events_on_repo_name` (`repo_name`),
    KEY                  `index_github_events_on_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
PARTITION BY LIST COLUMNS(`type`)
(PARTITION `push_event` VALUES IN ("PushEvent"),
 PARTITION `create_event` VALUES IN ("CreateEvent"),
 PARTITION `pull_request_event` VALUES IN ("PullRequestEvent"),
 PARTITION `watch_event` VALUES IN ("WatchEvent"),
 PARTITION `issue_comment_event` VALUES IN ("IssueCommentEvent"),
 PARTITION `issues_event` VALUES IN ("IssuesEvent"),
 PARTITION `delete_event` VALUES IN ("DeleteEvent"),
 PARTITION `fork_event` VALUES IN ("ForkEvent"),
 PARTITION `pull_request_review_comment_event` VALUES IN ("PullRequestReviewCommentEvent"),
 PARTITION `pull_request_review_event` VALUES IN ("PullRequestReviewEvent"),
 PARTITION `gollum_event` VALUES IN ("GollumEvent"),
 PARTITION `release_event` VALUES IN ("ReleaseEvent"),
 PARTITION `member_event` VALUES IN ("MemberEvent"),
 PARTITION `commit_comment_event` VALUES IN ("CommitCommentEvent"),
 PARTITION `public_event` VALUES IN ("PublicEvent"),
 PARTITION `gist_event` VALUES IN ("GistEvent"),
 PARTITION `follow_event` VALUES IN ("FollowEvent"),
 PARTITION `event` VALUES IN ("Event"),
 PARTITION `download_event` VALUES IN ("DownloadEvent"),
 PARTITION `team_add_event` VALUES IN ("TeamAddEvent"),
 PARTITION `fork_apply_event` VALUES IN ("ForkApplyEvent"));

CREATE TABLE `users`
(
    `id`           int(11) NOT NULL AUTO_INCREMENT,
    `login`        varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    `company`      varchar(255) COLLATE utf8mb4_unicode_ci          DEFAULT NULL,
    `company_name` varchar(255) COLLATE utf8mb4_unicode_ci          DEFAULT NULL,
    `created_at`   timestamp                               NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `type`         varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'USR',
    `fake`         tinyint(1) NOT NULL DEFAULT '0',
    `deleted`      tinyint(1) NOT NULL DEFAULT '0',
    `long`         decimal(11, 8)                                   DEFAULT NULL,
    `lat`          decimal(10, 8)                                   DEFAULT NULL,
    `country_code` char(3) COLLATE utf8mb4_unicode_ci               DEFAULT NULL,
    `state`        varchar(255) COLLATE utf8mb4_unicode_ci          DEFAULT NULL,
    `city`         varchar(255) COLLATE utf8mb4_unicode_ci          DEFAULT NULL,
    `location`     varchar(255) COLLATE utf8mb4_unicode_ci          DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY            `index_login_on_users` (`login`),
    KEY            `idx_company_name` (`company_name`),
    KEY            `users_cmp_idx` (`company`)
);

CREATE TABLE `github_repos`
(
    `repo_id`            int(11) NOT NULL,
    `repo_name`          varchar(150) NOT NULL,
    `owner_id`           int(11) NOT NULL,
    `owner_login`        varchar(255) NOT NULL,
    `owner_is_org`       tinyint(1) NOT NULL,
    `description`        varchar(512) NOT NULL DEFAULT '',
    `primary_language`   varchar(32)  NOT NULL DEFAULT '',
    `license`            varchar(32)  NOT NULL DEFAULT '',
    `size`               bigint(20) NOT NULL DEFAULT '0',
    `stars`              int(11) NOT NULL DEFAULT '0',
    `forks`              int(11) NOT NULL DEFAULT '0',
    `parent_repo_id`     int(11) DEFAULT NULL,
    `is_fork`            tinyint(1) NOT NULL DEFAULT '0',
    `is_archived`        tinyint(1) NOT NULL DEFAULT '0',
    `is_deleted`         tinyint(1) NOT NULL DEFAULT '0',
    `latest_released_at` timestamp NULL DEFAULT NULL,
    `pushed_at`          timestamp NULL DEFAULT NULL,
    `created_at`         timestamp    NOT NULL DEFAULT '0000-00-00 00:00:00',
    `updated_at`         timestamp    NOT NULL DEFAULT '0000-00-00 00:00:00',
    `last_event_at`      timestamp    NOT NULL DEFAULT '0000-00-00 00:00:00',
    `refreshed_at`       timestamp    NOT NULL DEFAULT '0000-00-00 00:00:00',
    PRIMARY KEY (`repo_id`) /*T![clustered_index] CLUSTERED */,
    KEY                  `index_gr_on_owner_id` (`owner_id`),
    KEY                  `index_gr_on_repo_name` (`repo_name`),
    KEY                  `index_gr_on_stars` (`stars`),
    KEY                  `index_gr_on_repo_id_repo_name` (`repo_id`,`repo_name`),
    KEY                  `index_gr_on_created_at_is_deleted` (`created_at`,`is_deleted`),
    KEY                  `index_gr_on_owner_login_owner_id_is_deleted` (`owner_login`,`owner_id`,`is_deleted`)
);


CREATE TABLE `blacklist_repos`
(
    `name` varchar(255) DEFAULT NULL
);

CREATE TABLE `blacklist_users`
(
    `login` varchar(255) NOT NULL,
    UNIQUE KEY `blacklist_users_login_uindex` (`login`),
    PRIMARY KEY (`login`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `collection_items`
(
    `id`            bigint(20) NOT NULL AUTO_INCREMENT,
    `collection_id` bigint(20) DEFAULT NULL,
    `repo_name`     varchar(255) NOT NULL,
    `repo_id`       bigint(20) NOT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY             `index_collection_items_on_collection_id` (`collection_id`)
);

CREATE TABLE `collections`
(
    `id`     bigint(20) NOT NULL AUTO_INCREMENT,
    `name`   varchar(255) NOT NULL,
    `public` tinyint(1) DEFAULT '1',
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    UNIQUE KEY `index_collections_on_name` (`name`)
);

CREATE TABLE `db_repos`
(
    `id`   varchar(255) NOT NULL,
    `name` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `cn_orgs`
(
    `id`      varchar(255) NOT NULL,
    `name`    varchar(255) DEFAULT NULL,
    `company` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `cn_repos`
(
    `id`      varchar(255) NOT NULL,
    `name`    varchar(255) DEFAULT NULL,
    `company` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `gh`
(
    `id`                 bigint(20) DEFAULT NULL,
    `type`               varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `created_at`         datetime                                DEFAULT NULL,
    `repo_id`            bigint(20) DEFAULT NULL,
    `repo_name`          varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `actor_id`           bigint(20) DEFAULT NULL,
    `actor_login`        varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `actor_location`     varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `language`           varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `additions`          bigint(20) DEFAULT NULL,
    `deletions`          bigint(20) DEFAULT NULL,
    `action`             varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `number`             int(11) DEFAULT NULL,
    `commit_id`          varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `comment_id`         bigint(20) DEFAULT NULL,
    `org_login`          varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `org_id`             bigint(20) DEFAULT NULL,
    `state`              varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `closed_at`          datetime                                DEFAULT NULL,
    `comments`           int(11) DEFAULT NULL,
    `pr_merged_at`       datetime                                DEFAULT NULL,
    `pr_merged`          tinyint(1) DEFAULT NULL,
    `pr_changed_files`   int(11) DEFAULT NULL,
    `pr_review_comments` int(11) DEFAULT NULL,
    `pr_or_issue_id`     bigint(20) DEFAULT NULL,
    `event_day`          date                                    DEFAULT NULL,
    `event_month`        date                                    DEFAULT NULL,
    `author_association` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `event_year`         int(11) DEFAULT NULL,
    `push_size`          int(11) DEFAULT NULL,
    `push_distinct_size` int(11) DEFAULT NULL,
    KEY                  `index_github_events_on_id` (`id`),
    KEY                  `index_github_events_on_action` (`action`),
    KEY                  `index_github_events_on_actor_id` (`actor_id`),
    KEY                  `index_github_events_on_actor_login` (`actor_login`),
    KEY                  `index_github_events_on_additions` (`additions`),
    KEY                  `index_github_events_on_closed_at` (`closed_at`),
    KEY                  `index_github_events_on_comment_id` (`comment_id`),
    KEY                  `index_github_events_on_comments` (`comments`),
    KEY                  `index_github_events_on_commit_id` (`commit_id`),
    KEY                  `index_github_events_on_created_at` (`created_at`),
    KEY                  `index_github_events_on_deletions` (`deletions`),
    KEY                  `index_github_events_on_event_day` (`event_day`),
    KEY                  `index_github_events_on_event_month` (`event_month`),
    KEY                  `index_github_events_on_event_year` (`event_year`),
    KEY                  `index_github_events_on_language` (`language`),
    KEY                  `index_github_events_on_org_id` (`org_id`),
    KEY                  `index_github_events_on_org_login` (`org_login`),
    KEY                  `index_github_events_on_pr_changed_files` (`pr_changed_files`),
    KEY                  `index_github_events_on_pr_merged_at` (`pr_merged_at`),
    KEY                  `index_github_events_on_pr_or_issue_id` (`pr_or_issue_id`),
    KEY                  `index_github_events_on_pr_review_comments` (`pr_review_comments`),
    KEY                  `index_github_events_on_repo_id` (`repo_id`),
    KEY                  `index_github_events_on_repo_name` (`repo_name`),
    KEY                  `index_github_events_on_type` (`type`)
);

CREATE TABLE `nocode_repos`
(
    `id`   varchar(255) NOT NULL,
    `name` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `web_framework_repos`
(
    `id`   varchar(255) NOT NULL,
    `name` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `programming_language_repos`
(
    `id`   varchar(255) NOT NULL,
    `name` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `static_site_generator_repos`
(
    `id`   varchar(255) NOT NULL,
    `name` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `js_framework_repos`
(
    `id`   varchar(255) NOT NULL,
    `name` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `css_framework_repos`
(
    `id`   varchar(255) NOT NULL,
    `name` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `trending_repos`
(
    `id`         bigint(20) NOT NULL AUTO_INCREMENT,
    `repo_name`  varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `created_at` datetime                                DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY          `index_trending_repos_on_repo_name` (`repo_name`),
    KEY          `index_trending_repos_on_created_at` (`created_at`)
);

ALTER TABLE db_repos SET TIFLASH REPLICA 1;
ALTER TABLE cn_orgs SET TIFLASH REPLICA 1;
ALTER TABLE cn_repos SET TIFLASH REPLICA 1;
ALTER TABLE users SET TIFLASH REPLICA 1;
ALTER TABLE gh SET TIFLASH REPLICA 1;
ALTER TABLE nocode_repos SET TIFLASH REPLICA 1;
ALTER TABLE web_framework_repos SET TIFLASH REPLICA 1;
ALTER TABLE programming_language_repos SET TIFLASH REPLICA 1;
ALTER TABLE static_site_generator_repos SET TIFLASH REPLICA 1;
ALTER TABLE js_framework_repos SET TIFLASH REPLICA 1;
ALTER TABLE css_framework_repos SET TIFLASH REPLICA 1;
ALTER TABLE github_events SET TIFLASH REPLICA 1;
ALTER TABLE blacklist_users SET TIFLASH REPLICA 1;
ALTER TABLE github_repos SET TIFLASH REPLICA 1;
-- ALTER TABLE csdn_events SET TIFLASH REPLICA 1;
ALTER TABLE trending_repos SET TIFLASH REPLICA 1;