from doctest import register_optionflag
import requests
import sys

URL = "http://localhost:3000"

store = {
    "access": "",
    "refresh": "",
    "old_refresh": ""
}


def register():
    content = {
        "Username": "johndoe",
        "Password": "12345",
        "Firstname" : "John",
        "Lastname" : "Doe",
        "Email": "jogn.doe@example.com",
        "Country": "US"
    }
    res = requests.post(f"{URL}/register", json=content)
    try:
        assert res.status_code == 201
    except AssertionError:
        print(f"Register Failed: {res.text}")
        sys.exit(1)


def login():
    content = {
        "username": "johndoe",
        "password": "12345"
    }
    res = requests.post(f"{URL}/login", json=content)
    try:
        assert res.status_code == 201
    except:
        print(f"Login failed: {res.text}")
        sys.exit(1)

    data = res.json()
    store['access'] = data["AccessToken"]
    store['refresh'] = data["RefreshToken"]


def testToken():
    headers = {
        'Authorization': f'Bearer {store["access"]}'
    }
    res = requests.get(f"{URL}/checkjwt", headers=headers)
    try:
        assert res.status_code == 200
    except:
        print(f"Token test faild: {res.text}")
        sys.exit(1)


def refreshToken():

    headers = {
        'Authorization': f'Bearer {store["access"]}'
    }
    body = {"refresh_token": f"{store['refresh']}"}

    res = requests.post(f"{URL}/refresh", headers=headers, json=body)
    try:
        assert res.status_code == 201
        data = res.json()
        store["access"] = data["AccessToken"]
        store["refresh"] = data["RefreshToken"]
    except:
        print(f"Refresh Test Failed: {res.text}")
        sys.exit(1)
    

def refreshDoubleUse():
    store["old_refresh"] = store["refresh"]

    refreshToken()

    headers ={"Authorization": f"Bearer {store['access']}"}
    body = {"refresh_token": f"{store['old_refresh']}"}

    res = requests.post(f"{URL}/refresh", headers=headers, json=body)
    try:
        assert res.status_code == 401
    except:
        print("Refresh with INVALID token did not work as expected")
        print(f"Error if applicable: {res.text}")
        sys.exit(1)

    body = {"refresh_token": f"{store['refresh']}"}

    res = requests.post(f"{URL}/refresh", headers=headers, json=body)
    
    try:
        assert res.status_code == 401
    except:
        print("Refresh with DELETED token did not work as expected")
        print(f"Error if applicable: {res.text}")
        sys.exit(1)



if __name__ == "__main__":
    register()

    login()
    testToken()

    refreshToken()
    testToken()

    refreshDoubleUse() # all sessions deleted

    login()
    refreshToken()
    

    print('\ntests complete')

