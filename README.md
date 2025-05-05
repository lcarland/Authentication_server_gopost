Authentication Server with Postgres
===================================

Note: This is not intended to be a production ready server and is for demonstrative purposes only.

What's Featured
---------------
- JWT Authentication
    - Access tokens are short-lived to minimize the impact of token theft
- Refresh Tokens to renew expired JWT
    - Refresh tokens are rotated after one use
    - Tokens are saved in their own table, so a user can have multiple refresh tokens for different clients
    - Double use of refresh tokens is monitored and results in all refresh tokens becoming invalidated
- Forgotten password reset
- Create, delete, and modify user records
- Basic privileges implemented
    - a non-staff user cannot modify another user's info
    - a user can only view certain info related to another user
    - staff can modify, delete, and view all info related to other users

What's Missing
--------------
- Email Service for password resetting. Reset tokens are sent in the response body for testing purposes.
- Email verification on sign-up
- Multi-Factor authentication and/or clinet fingerprinting

Setup
-----
Generate an RSA Key pair in PEM format for singing JWTs.

A .env file is required in the project root.
The file should contain these variables:
- POSTGRES_USER=*pgadmin_user*
- POSTGRES_PASSWORD=*Long Acsii String*  

*Secret key for password hashing*
- SECRET_KEY=*Long Acsii String*

*Port numbers for Postgres and Go Application*
- PG_PORT=*3000*
- GO_PORT=*5432*

*ED25519 Key pair paths for JWT signing. Files should be PEM format and either absolute or relative to the main.go file.*
- PRIV_KEY=*./private_key.pem*
- PUB_KEY=*./public_key.pem*

<br><br>

API Reference
=======================================================
Annotations
-------------------------------------------------------
@TokenRequired:  
Access token required  
Headers = "Authorization": "Bearer *${JSON-Web-Token}* "

@CredentialsRequired:  
username and password required  
JSON:
```
{
    "Username": string,
    "Password": string
}
```

Routes
-------------------------------------------------------
Overview:
```
/                   GET
/{country}          GET
/user               POST
/user/{id}          GET, PATCH, DELETE
/user/password      POST, PUT
/session            POST, DELETE
/session/refresh    POST
/checkjwt           GET
/publickey          GET
```

/
---
GET -> JSON

Display all country code information
```
[
    {
        "code": "GB",
        "country": "United Kingdom",
        "dialcode":"+44"
    },
    {
        "code": "US",
        "country": "United States",
        "dialcode":"+1"
    },
    ....
]
```

/{country}
----------
GET -> JSON

Get country info based on ISO Code
```
GET "/US"
{
    "code": "US",
    "country": "United States",
    "dialcode":"+1"
}
```

/user
-----
POST: JSON -> 201

Register new user
```
request_body:
{
    "username": string,
    "password": string,
    "first_name": string,
    "last_name": string,
    "email": string,
    "phone": string || null,
    "country": string || null
}
```

/user/{id}
----------
@TokenRequired  
GET -> JSON

Get user information. Public and private user info is sent depending to permission

```
response:
{
    "id": int,
    "username": string,
    "country": string,
    "date_joined": datetime,
    "is_activte": bool,

    // Private data included if user is self or staff
    "email": string,
    "phone": string,
    "is_superuser": bool,
    "is_staff": bool,
    "last_login": datetime
}
```

@TokenRequired  
PATCH: JSON -> 200  

Update user information

```
request_body:
{
    // any of the following fields
    "username": string,
    "first_name": string,
    "last_name": string,
    "email": string,
    "phone": string,
    "country": string
}
```

@TokenRequired  
@CredentialsRequired  
DELETE: JSON -> 204  

Permanently removes user account

/user/password
--------------
POST: JSON -> 201

Lost password. Generates token to reset password. Lasts 5 minutes.
```
request_body:
{
    "email": string
}
```

PUT: JSON -> 202

Change password with token
```
request_body:
{
    "token": string,
    "username": string,
    "password": string  // new password
}
```

/session
--------
@CredentialsRequired  
POST: JSON -> JSON

Login user
```
response:
{
    "AccessToken": string,
    "RefreshToken": string
}
```

@TokenRequired  
DELETE -> 204

Essentially logs out user by deleting Refresh Token. Client is responsible for deleting access and refresh tokens.

/session/refresh
----------------
@TokenRequired  
POST: JSON -> JSON  

Refreshes access token and rotates refresh token
```
request_body:
{
    "refresh_token": string
}

response:
{
    "AccessToken": string,
    "RefreshToken": string
}
```

/checkjwt
---------
@TokenRequired  
GET -> 200

Confirms that JWT is valid. Mostly for testing purposes.

/publickey
----------
GET -> Text

Returns Public Key in PEM Fromat
