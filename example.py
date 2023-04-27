import json
import requests

addr = 'http://localhost:8080'

def create_wallet(name, idempotency_key='XXXXYYYY'):
    print("create {} with ik {}".format(name, idempotency_key))

    payload=json.dumps({'wallet_name': name})

    headers = {
      'Idempotency-Key': idempotency_key,
      'Content-Type': 'application/json'
    }

    response = requests.request('POST', addr + '/wallet', headers=headers, data=payload)

    if response.status_code == 200:
        print('\t{} created'.format(name))
    elif response.status_code == 201:
        print('\t{} already created'.format(name))
    elif response.status_code == 409:
        print('\tcannot create because name duplication {}'.format(name), response.json())
    else:
        print('\tcannot create', response.status_code, response.json())

def wallet(name):
    print('wallet info {}'.format(name))

    response = requests.request("GET", addr + '/wallet/' + name)
    print(json.dumps(response.json(), indent=4, sort_keys=True))

def deposit(name, amount, idempotency_key='XXXXYYYY'):

    print('deposit {} to wallet {} with ik {}'.format(amount, name, idempotency_key))

    payload=json.dumps({'wallet_name': name, 'amount': str(amount)})
    headers = {
      'Idempotency-Key': idempotency_key,
      'Content-Type': 'application/json'
    }

    response = requests.request('POST', addr + '/deposit', headers=headers, data=payload)
    if response.status_code == 200:
        print('\tdone')
    elif response.status_code == 201:
        print('\talready done')
    else:
        print('\tcannot make', response.status_code, response.json())


def transfer(name_from, name_to, amount, idempotency_key='XXXXYYYY'):

    print('transfer {} from {} to {} with ik {}'.format(amount, name_from, name_to, idempotency_key))

    payload=json.dumps({'wallet_name_from': name_from, 'wallet_name_to': name_to, 'amount': str(amount)})
    headers = {
      'Idempotency-Key': idempotency_key,
      'Content-Type': 'application/json'
    }

    response = requests.request('POST', addr + '/transfer', headers=headers, data=payload)
    if response.status_code == 200:
        print('\tdone')
    elif response.status_code == 201:
        print('\talready done')
    else:
        print('\tcannot make', response.status_code, response.json())


def history(name):
    print('history for wallet {}'.format(name))

    response = requests.request("GET", addr + '/history/' + name)
    print(response.text)
    # print(json.dumps(response.json(), indent=4, sort_keys=True))

wallet('myWallet')  # not found or ok
wallet('anotherWallet')  # not found or ok

create_wallet('myWallet', idempotency_key='create_key_1')  # ok
create_wallet('myWallet', idempotency_key='create_key_1')  # idempotency_key duplication

create_wallet('anotherWallet', idempotency_key='create_key_2')  # ok
create_wallet('anotherWallet', idempotency_key='create_key_2')  # idempotency_key duplication

create_wallet('myWallet', idempotency_key='create_key_3')  # duplication in database
create_wallet('anotherWallet', idempotency_key='create_key_4')  # duplication in database

deposit('myWallet', 10000, idempotency_key='deposit_1')  # ok
deposit('myWallet', 10000, idempotency_key='deposit_1')  # idempotency_key duplication

wallet('myWallet')  # ok
wallet('anotherWallet')  # ok

for i in range(3):
    transfer('myWallet', 'anotherWallet', 100, 'transfer_' + str(i))  # ok or idempotency_key duplication

for i in range(3):
    transfer('anotherWallet', 'myWallet', 50, 'transfer_' + str(10 + i))  # ok or idempotency_key duplication

transfer('myWallet', 'anotherWallet', 100000, 'transfer_1000')  # insufficiet balance

history('myWallet')  # 7(deposit + 6 transfers) items for myWallet
history('anotherWallet')  # 6 transfer items for anotherWallet
