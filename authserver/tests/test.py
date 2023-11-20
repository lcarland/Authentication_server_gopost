import requests

URL = "http://localhost:3000"

res = requests.get(URL)
assert type(res.json()) == type(list())

res = requests.get(f"{URL}/US")
assert type(res.json()) == type(dict())

print("tests passed")
