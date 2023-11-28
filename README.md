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
...and I'm sure more

Setup
-----
A .env file is required in the project root.
The file should contain these variables:
- POSTGRES_USER
- POSTGRES_PASSWORD
- SECRET_KEY
- ACCESS_KEY
