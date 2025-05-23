import requests
import sys


URL = "http://localhost:3000"


store = {
    "access": "",
    "refresh": "",
    "old_refresh": "",
    "user_url": "",
    "reset": ""
}

HELPTXT = """
Help:

Useage:
    -h help - prints this help text
    -r register user
    -pw tests with change user password
    -pk gets the public key
    none - perform the other tests

"""


user1 = {
    "Username": "johndoe",
    "Password": "12345",
    "Firstname" : "John",
    "Lastname" : "Doe",
    "Email": "john.doe@example.com",
    "Country": "US"
}

user2 = {
    "Username": "cedardog",
    "Password": "1534ghtk",
    "Firstname": "Cedar",
    "Lastname": "Dog",
    "Email": "cedardog@barkmail.com",
    "Country": "XX"
}

PW_UPDATE = "f88hfhhs2"


def register():
    res = requests.post(f"{URL}/user", json=user1)
    try:
        assert res.status_code == 201
    except AssertionError:
        print(f"Register Failed: {res.text}")
        sys.exit(1)


def login():
    content = {
        "username": user1["Username"],
        "password": user1["Password"]
    }
    res = requests.post(f"{URL}/session", json=content)
    try:
        assert res.status_code == 201
    except AssertionError:
        print(f"Login failed: {res.text}")
        sys.exit(1)

    store['user_url'] = res.headers['Content-Location']
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
    except AssertionError:
        print(f"Token test failed: {res.text}")
        sys.exit(1)


def testNoToken():
    res = requests.get(f"{URL}/checkjwt")
    try:
        assert res.status_code == 400
    except AssertionError:
        print(f"No Token in Header test fail. {res.text}")
        sys.exit(1)


def refreshToken():

    headers = {
        'Authorization': f'Bearer {store["access"]}'
    }
    body = {"refresh_token": f"{store['refresh']}"}

    res = requests.post(f"{URL}/session/refresh", headers=headers, json=body)
    try:
        assert res.status_code == 201
        data = res.json()
        store["access"] = data["AccessToken"]
        store["refresh"] = data["RefreshToken"]
    except AssertionError:
        print(f"Refresh Test Failed: {res.text}")
        sys.exit(1)
    

def refreshDoubleUse():
    """Move current refresh token to old_refresh in the data store. Token refresh
    is called again to invalidate token in old_refresh. Refresh route is ran with 
    old token to delete all refresh tokens. This is checked with another refresh 
    attempt with the 'good' token, which should fail."""

    store["old_refresh"] = store["refresh"]

    refreshToken()

    headers ={"Authorization": f"Bearer {store['access']}"}
    body = {"refresh_token": f"{store['old_refresh']}"}

    res = requests.post(f"{URL}/session/refresh", headers=headers, json=body)
    try:
        assert res.status_code == 401
    except:
        print("Refresh with INVALID token did not work as expected")
        print(f"Error if applicable: {res.text}")
        sys.exit(1)

    body = {"refresh_token": f"{store['refresh']}"}

    res = requests.post(f"{URL}/session/refresh", headers=headers, json=body)
    
    try:
        assert res.status_code == 401
    except AssertionError:
        print("Refresh with DELETED token did not work as expected")
        print(f"Error if applicable: {res.text}")
        sys.exit(1)


def get_user():
    headers = {"Authorization": f"Bearer {store['access']}"}
    res = requests.get(f"{URL}{store['user_url']}", headers=headers)
    try:
        assert res.status_code == 200
    except AssertionError:
        print(f"Get User Failed: {res.text}")
        sys.exit(1)


def update_profile():
    content = {
        "email": "johndoe@newemail.com"
    }
    headers = {"Authorization": f"Bearer {store['access']}"}
    res = requests.patch(f"{URL}{store['user_url']}", headers=headers, json=content)
    try:
        assert res.status_code == 200
    except AssertionError:
        print(f"Update Profile Fail: {res.text}")
        sys.exit(1)


def password_reset_init():
    content = {"email": "johndoe@newemail.com"}
    res = requests.post(f"{URL}/user/password", json=content)
    try:
        assert res.status_code == 201
        data = res.json()
    except AssertionError:
        print(f"Password reset request failed: {res.text}")
        sys.exit(1)

    store["reset"] = data['reset_token']


def password_reset(new_pw: str):
    content = {
        "token": store["reset"],
        "username": "johndoe",
        "password": new_pw
    }
    res = requests.put(f"{URL}/user/password", json=content)
    try: 
        assert res.status_code == 202
    except AssertionError:
        print(f"Password reset failed: {res.text}")
        sys.exit(1)
    
    user1["Password"] = content["password"]


def get_pub_key():
    res = requests.get(f"{URL}/publickey")
    print(res.text)



### User 2 functions
def register2():
    res = requests.post(f"{URL}/user", json=user2)

    try:
        assert res.status_code == 201
    except AssertionError:
        print(f"register2 failed: {res.text}")
        sys.exit(1)


if __name__ == "__main__":
    arg = sys.argv

    chgpw = False

    if len(arg) > 1:
        if arg[1] == '-r':
            register()
        elif arg[1] == "-pw":
            chgpw = True
        elif arg[1] == "-pk":
            get_pub_key()
            sys.exit(0)
        elif arg[1] == '-h':
            print(HELPTXT)
            sys.exit(0)
        else:
            print("invalid option: use -h for help")
            sys.exit(0)

    # General login test
    # Test with and without token
    login()
    testToken()
    testNoToken()

    # refresh test
    # ensure new token takes
    refreshToken()
    testToken()

    # refresh with previously used token
    refreshDoubleUse() # all sessions deleted

    # login after refresh double use
    login()
    refreshToken()

    get_user()
    update_profile()

    if chgpw:
        password_reset_init()
        password_reset(PW_UPDATE)
        print("Password changed\n")
        password_reset_init()
        password_reset(user1["Password"])
        print("Password changed back to original\n")

    print('\nAll Tests Passed!')

