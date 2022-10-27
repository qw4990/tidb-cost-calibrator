# https://github.com/pingcap/ossinsight/tree/main/api/queries
import urllib.request
import json

def gen2(idx, name, sql, params):
    if idx == len(params):
        fp = open(name+'.sql', 'w')
        fp.write(sql)
        fp.close()
        return
    
    if 'enums' not in params[idx]:
        gen2(idx+1, name, sql, params)
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

def readurl(url):
    fp = urllib.request.urlopen(url)
    mybytes = fp.read()
    return mybytes.decode("utf8")

name = "stars-top-50-company"
sql_url = ('https://raw.githubusercontent.com/pingcap/ossinsight/main/api/queries/%s/template.sql' % name)
param_url = ('https://raw.githubusercontent.com/pingcap/ossinsight/main/api/queries/%s/params.json' % name)

print("read template.sql")
template_sql = readurl(sql_url)
print(template_sql)
print("==================================================================")
print("read params.json")
params_json = readurl(param_url)
print(params_json)
print("==================================================================")
gen(name, template_sql, params_json)