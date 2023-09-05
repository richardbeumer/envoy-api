import logging
import requests
import os
import sys
from logger import Logger
from fastapi import FastAPI
from bs4 import BeautifulSoup

user=os.getenv('ENLIGHTEN_USERNAME')
password=os.getenv('ENLIGHTEN_PASSWORD')
envoy_serial=os.getenv('ENVOY_SERIAL')
envoy_site=os.getenv('ENVOY_SITE')
envoy_host=os.getenv('ENVOY_HOST')
LOGIN_URL = "https://entrez.enphaseenergy.com/login"
TOKEN_URL = "https://entrez.enphaseenergy.com/entrez_tokens"
token = ''

app = FastAPI()
logger = Logger.get_logger(__name__)
logger.info('API Started')
    
def get_token():
    payload_login = {
        "username": user,
        "password": password,
    }
    login_response = requests.post(LOGIN_URL, data=payload_login)
    payload_token = {
        "Site": envoy_site,
        "serialNum": envoy_serial,
    }
    token_response = requests.post(TOKEN_URL, data=payload_token, cookies=login_response.cookies)
    parsed_html = BeautifulSoup(token_response.text, features="html.parser")
    token = parsed_html.body.find(  # pylint: disable=invalid-name, unused-variable, redefined-outer-name
        "textarea"
    ).text
    return token

def check_token(token):
    requests.packages.urllib3.disable_warnings()
    endpoint = 'https://' + envoy_host + '/auth/check_jwt'
    headers = {"Content-Type":"application/json", "Authorization": f"Bearer {token}"}
    response = requests.get(url=endpoint, headers=headers, verify=False)
    
    if "Valid token." not in response.text:
        logger.info('Refreshing token')
        return (get_token())
        
    return token


@app.get("/production/")
def get_envoy_data():
    requests.packages.urllib3.disable_warnings()
    global token
    token = check_token(token)
    headers = {"Content-Type":"application/json", "Authorization": f"Bearer {token}"}
    endpoint = 'https://' + envoy_host + '/ivp/pdm/energy'
    data=requests.get(url=endpoint, headers=headers, verify=False).json()
 
    return data['production']['pcu']
