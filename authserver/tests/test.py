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
    except AssertionError:
        print(f"Token test faild: {res.text}")
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


def update_profile():
    content = {
        "email": "johndoe@newemail.com"
    }
    headers = {"Authorization": f"Bearer {store['access']}"}
    res = requests.patch(f"{URL}/user/")


def clean_db():
    res = requests.delete(f"{URL}/cleanusers")
    print(f"{res.status_code} {res.text}")


if __name__ == "__main__":
    arg = sys.argv
    if len(arg) > 1:
            if arg[1] == '-r':
                register()
            elif arg[1] == '-c':
                clean_db()
                sys.exit(0)
            elif arg[1] == '-h':
                helpstr = """
Help:

Useage:
    -h help - prints this help text
    -r register user
    -c delete user entries
    none - perform the other tests
                """ 
                print(helpstr)
                sys.exit(0)
            else:
                print("invalid option: use -h for help")
                sys.exit(0)

    # General login test
    login()
    testToken()
    testNoToken()

    # refresh test
    refreshToken()
    testToken()

    # refresh with previously used token
    refreshDoubleUse() # all sessions deleted

    # login after refresh double use
    login()
    refreshToken()


    

    print('\ntests complete')

