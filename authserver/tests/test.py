import requests

URL = "http://localhost:3000"

store = {
    "access": "",
    "refresh": ""
}


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
        return

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
        return


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
        return


if __name__ == "__main__":
    login()
    testToken()

    refreshToken()
    testToken()

    print('\ntests complete')

