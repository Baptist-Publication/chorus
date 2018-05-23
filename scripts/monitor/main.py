import json

import datetime
import requests
import time
from jinja2 import Template, Environment, FileSystemLoader

import sendmail

RPC_PORT = 46657

s = requests.Session()

# list of emails
receivers = []


def epoch_ftime(epoch):
    return datetime.datetime.fromtimestamp(epoch).strftime('%Y-%m-%d %H:%M:%S')


def check_height(params):
    ip = params['IP']

    resp = s.get('http://%s:%d/status' % (ip, RPC_PORT))
    j = json.loads(resp.text)
    params['node_start'] = epoch_ftime(int(j['result']['node_info']['other'][2].split('=')[1]))
    params['latest_height'] = j['result']['latest_block_height']
    params['latest_block_time'] = epoch_ftime(j['result']['latest_block_time'] / 1000000000)
    params['revision'] = j['result']['node_info']['other'][3].split('=')[1][:10]

    params['pubkey'] = j['result']['node_info']['pub_key'][1]


def check_is_validator(params):
    ip = params['IP']

    # resp = s.get('http://%s:%d/is_validator?pubkey=0x%s' % (ip, RPC_PORT, params['pubkey']))
    # j = json.loads(resp.text)

    resp = s.get('http://%s:%d/validators' % (ip, RPC_PORT))
    j = json.loads(resp.text)
    pub_keys = [x['pub_key'][1] for x in j['result']['validators']]
    params['validator'] = params['pubkey'] in pub_keys


def handle(ip):
    params = {'IP': ip}
    try:
        check_height(params)
        check_is_validator(params)
        params['ok'] = True
    except Exception, err:
        print err
        params['ok'] = False

    return params


def render_report(all_results):
    headers = ['ok', 'validator', 'latest_height', 'node_start', 'latest_block_time']
    ok_len = len([x['ok'] for x in all_results])

    max_height = max([x['latest_height'] for x in all_results])
    min_height = min([x['latest_height'] for x in all_results])

    overall = {'time': epoch_ftime(time.time())}

    if ok_len == len(all_results) and max_height - min_height <= 1:
        overall['status'] = 'ALL GOOD. HEIGHT ALIGNED.'
        overall['status_color'] = '#62f442'
    elif ok_len == 0:
        overall['status'] = 'ALL DOWN'
        overall['status_color'] = '#ff3030'
    else:
        overall['status'] = 'NOT GOOD: alive: %d Max Height: %d Min Height: %d' % (ok_len, max_height, min_height)
        overall['status_color'] = '#ffd456'

    env = Environment(loader=FileSystemLoader(searchpath='.'), trim_blocks=True, lstrip_blocks=True)
    template = env.get_template('report.template.html')
    title = 'Chorus: %s %s' % (overall['status'], overall['time'])
    html = template.render(headers=headers, lines=all_results, overall=overall)
    return title, html


if __name__ == '__main__':
    t = int(time.time())
    all_results = []
    with open('ips') as f:
        for line in f:
            params = handle(line.strip())
            all_results.append(params)
            print params
            if not params['ok']:
                with open('notok_%d.log' % (t), 'a') as f:
                    print >> f, params['IP']
            # break
    all_results = sorted(all_results, key=lambda x: x["IP"])
    title, html = render_report(all_results)

    sendmail.send(receivers, title, html)
