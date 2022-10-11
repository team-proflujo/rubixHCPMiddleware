# Script to create Registration token in HCP
import requests, traceback, json, os, argparse
from typing import Final

CONFIG_JSON_FILE_PATH: Final = 'config.json'

# main
def main():
    try:
        parser = argparse.ArgumentParser()
        parser.add_argument('-t', '--token', help = 'Admin Token from HCP Vault')
        args = parser.parse_args()

        if not args.token:
            print('Invalid Arguments! token argument must be provided.')
            quit()

        if os.path.isfile(CONFIG_JSON_FILE_PATH):
            configData = None

            with open(CONFIG_JSON_FILE_PATH, 'r') as fpConfigJson:
                try:
                    configData = json.loads(fpConfigJson.read())
                except Exception as e:
                    print(f'Error when parsing {CONFIG_JSON_FILE_PATH}')
                    traceback.print_exc()
                    quit()

                if configData and type(configData) is dict:
                    if 'hcpVaultStorageConfig' in configData and 'apiURL' in configData['hcpVaultStorageConfig'] and 'namespace' in configData['hcpVaultStorageConfig'] and 'secretEngineName' in configData['hcpVaultStorageConfig']:
                        if configData['hcpVaultStorageConfig']['apiURL'] and configData['hcpVaultStorageConfig']['namespace'] and configData['hcpVaultStorageConfig']['secretEngineName']:

                            response = requests.post(configData['hcpVaultStorageConfig']['apiURL'] + '/v1/auth/token/create', headers = {
                                'X-Vault-Token': args.token,
                                'X-Vault-Namespace': configData['hcpVaultStorageConfig']['namespace']
                            }, data = {
                                'policies': ['register-wallet'],
                                'no_parent': True,
                                'period': '24h'
                            })

                            vaultResponse = None

                            if response and response.status_code != 204:
                                try:
                                    vaultResponse = response.json()
                                except Exception as e:
                                    print('Error when parsing response from Vault:')
                                    traceback.print_exc()
                                    quit()

                            if vaultResponse and type(vaultResponse) is dict:
                                if 'auth' in vaultResponse and vaultResponse['auth'] and 'client_token' in vaultResponse['auth'] and vaultResponse['auth']['client_token']:
                                    print('Register Token: ' + vaultResponse['auth']['client_token'])
                                else:
                                    print('Invalid response from Vault:')
                                    print(vaultResponse)
                            else:
                                print('Invalid response from Vault:')
                                print(vaultResponse)

                        else:
                            print('Invalid config data!')
                    else:
                        print('Invalid config data!')
                else:
                    print('Invalid config data!')
        else:
            print(f'{CONFIG_JSON_FILE_PATH} does not exists!')
    except Exception as e:
        print("Error occurred:")
        traceback.print_exc()
# main

if __name__ == '__main__':
    main()
