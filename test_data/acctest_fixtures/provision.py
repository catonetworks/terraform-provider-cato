#!/usr/bin/env python3
"""Provision persistent public-API fixtures; keep credentials and state outside Git."""

import argparse
import concurrent.futures
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import urllib.request


def graphql(account, endpoint, query, variables=None):
    request = urllib.request.Request(endpoint, json.dumps({
        'query': query, 'variables': dict(variables or {}, id=account['id']),
    }).encode(), {'Content-Type': 'application/json', 'x-api-key': account['api_key']})
    with urllib.request.urlopen(request, timeout=60) as response:
        result = json.load(response)
    if result.get('errors'):
        # Do not print response bodies or variables that could include credentials.
        raise RuntimeError('Fixture API request failed: ' + ', '.join(
            str(error.get('message', 'unknown error')) for error in result['errors']))
    return result['data']


def provision_users(account, endpoint):
    data = graphql(account, endpoint,
                   'query($id:ID!){user(accountId:$id){userList(input:{}){items{id email userStatus}}}}')
    users = {user['email']: user for user in data['user']['userList']['items']}
    fixture_ids = []
    for index in range(3):
        email = f'tf-fixture-{index}@example.invalid'
        if email not in users:
            result = graphql(account, endpoint,
                'mutation($id:ID!,$input:CreateUserInput!){user(accountId:$id){createUser(input:$input){user{id}}}}',
                {'input': {'firstName': 'TF Fixture', 'lastName': str(index), 'email': email}})
            users[email] = result['user']['createUser']['user']
        user_id = int(users[email]['id'])
        fixture_ids.append(str(user_id))
        # Fresh accounts allocate ID 0, which the current Terraform user resource
        # rejects. The API accepts it, so bootstrap these users directly.
        graphql(account, endpoint,
            'mutation($id:ID!,$input:EnableUserInput!){user(accountId:$id){enableUser(input:$input){users{id userStatus}}}}',
            {'input': {'userId': [user_id]}})
    configuration = graphql(account, endpoint,
        'query($id:ID!){access(accountId:$id){remoteAccessEligibility{allocationType '
        'entityList(input:{paging:{limit:1000,from:0}}){items{id type}}}}}')['access']['remoteAccessEligibility']
    existing = configuration['entityList']['items']
    if len(existing) >= 1000:
        raise RuntimeError('Eligibility list needs pagination before it can be updated safely')
    if configuration['allocationType'] == 'ALL' and len(users) > 3:
        raise RuntimeError('Review eligibility before bootstrapping a populated account')
    selected = {('user' if item['type'] == 'USER' else 'userGroup', item['id']) for item in existing}
    selected.update(('user', user_id) for user_id in fixture_ids)
    graphql(account, endpoint,
        'mutation($id:ID!,$input:AccessRemoteAccessSetEligibilityListInput!){access(accountId:$id){'
        'remoteAccessEligibility{setEligibilityList(input:$input){status}}}}',
        {'input': {'allocationType': 'CUSTOM', 'items': [
            {kind: {'by': 'ID', 'input': ref}} for kind, ref in sorted(selected)]}})


def provision_webhooks(account, endpoint):
    data = graphql(account, endpoint,
        'query($id:ID!){entityLookup(accountID:$id,type:webhookSubscription){items{entity{id name}}}}')
    existing = {item['entity']['name'] for item in data['entityLookup']['items']}
    for index in range(3):
        name = f'tf-fixture-webhook-{index}'
        if name in existing:
            continue
        # Disabled, unattached and addressed to a reserved invalid domain. Never
        # invoke testWebhook: fixture creation must not send external requests.
        graphql(account, endpoint,
            'mutation($id:ID!,$input:CreateWebhookInput!){notification(accountId:$id){integration{'
            'createWebhook(input:$input){webhook{id}}}}}',
            {'input': {'name': name, 'enabled': False, 'url': 'https://tf-fixture.example.invalid/webhook',
                       'auth': {}, 'body': '{}'}})


def provision_advanced_groups(account, endpoint):
    data = graphql(account, endpoint,
        'query($id:ID!){groups(accountId:$id){groupList(input:{paging:{limit:100,from:0},'
        'filter:[{name:[{regex:"^acctest_advanced_group_00[0-2]$"}]}],'
        'sort:{name:{direction:ASC,priority:1}}}){items{id name}}}}')
    existing = {group['name'] for group in data['groups']['groupList']['items']}
    for index in range(3):
        name = f'acctest_advanced_group_{index:03d}'
        if name not in existing:
            # Match the acceptance helper's persistent names. Select only the
            # identity, so no generated SDK members-projection arguments leak in.
            graphql(account, endpoint,
                'mutation($id:ID!,$input:CreateGroupInput!){groups(accountId:$id){'
                'createGroup(input:$input){group{id name}}}}',
                {'input': {'name': name, 'description': name + ' terraform tests'}})


def provision(account, args, state_root):
    directory = state_root / account['id']
    directory.mkdir(mode=0o700, parents=True, exist_ok=True)
    shutil.copyfile(Path(__file__).with_name('main.tf'), directory / 'main.tf')
    env = dict(os.environ, CATO_TOKEN=account['api_key'], CATO_BASEURL=args.endpoint,
               TF_VAR_account_id=account['id'], TF_CLI_CONFIG_FILE=str(state_root / 'terraform.rc'),
               DISABLE_POLICY_RULE_CLEANUP='true')
    env.pop('TF_LOG', None)
    env.pop('TF_LOG_PROVIDER', None)
    env.pop('TF_LOG_PATH', None)
    with (directory / 'apply.log').open('w') as log:
        result = subprocess.run(['terraform', f'-chdir={directory}', 'apply', '-auto-approve',
                                 '-input=false', '-parallelism=1', '-no-color'],
                                env=env, stdout=log, stderr=subprocess.STDOUT, check=False)
    if result.returncode:
        raise RuntimeError(f'Fixture apply failed for account {account["id"]}; see {directory / "apply.log"}')
    result = subprocess.run(['terraform', f'-chdir={directory}', 'output', '-json', 'test_vars'],
                            env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True, text=True)
    # Development-override warnings are on stderr; output JSON contains references only.
    test_vars = json.loads(result.stdout)
    provision_users(account, args.endpoint)
    provision_webhooks(account, args.endpoint)
    provision_advanced_groups(account, args.endpoint)
    return account['id'], test_vars


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--accounts', type=Path, required=True)
    parser.add_argument('--state-dir', type=Path, required=True)
    parser.add_argument('--provider-dir', type=Path, required=True)
    parser.add_argument('--endpoint', required=True)
    args = parser.parse_args()
    if not args.endpoint.startswith('https://'):
        parser.error('--endpoint must use HTTPS')
    os.umask(0o077)
    accounts = json.loads(args.accounts.read_text())
    if not isinstance(accounts, list) or not accounts:
        parser.error('--accounts must contain a nonempty JSON array')
    for account in accounts:
        if (not isinstance(account, dict) or not isinstance(account.get('id'), str)
                or not re.fullmatch(r'[0-9]+', account['id'])
                or not isinstance(account.get('api_key'), str) or not account['api_key']):
            parser.error('Every account needs a numeric string id and a nonempty api_key')
    if len({account['id'] for account in accounts}) != len(accounts):
        parser.error('Account IDs must be unique')
    state_root = args.state_dir.resolve()
    state_root.mkdir(mode=0o700, parents=True, exist_ok=True)
    # JSON string encoding is also valid for these HCL quoted strings.
    provider_dir = json.dumps(str(args.provider_dir.resolve()))
    (state_root / 'terraform.rc').write_text(
        'provider_installation {\n dev_overrides {\n "catonetworks/cato" = ' + provider_dir + '\n }\n direct {}\n}\n')
    with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
        futures = {pool.submit(provision, account, args, state_root): account['id'] for account in accounts}
        for future in concurrent.futures.as_completed(futures):
            account_id, test_vars = future.result()
            for account in accounts:
                if account['id'] == account_id:
                    account.setdefault('test_vars', {}).update(test_vars)
            args.accounts.write_text(json.dumps(accounts, indent=2) + '\n')
            args.accounts.chmod(0o600)
            print(f'Provisioned public fixtures for account {account_id}', flush=True)


if __name__ == '__main__':
    main()
