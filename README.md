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

*RSA Key pair paths for JWT signing. Files should be PEM format and either absolute or relative to the main.go file.*
- PRIV_KEY=*./private_key.pem*
- PUB_KEY=*./public_key.pem*
