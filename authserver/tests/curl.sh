curl -X POST http://localhost:3000/register --verbose \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d '{
        "Username": "johndoe",
        "FirstName": "John",
        "LastName": "Doe", "Country": "US", "Password":"12345",
        "Email": "john.doe@example.com"
    }'

curl http://localhost:3000/login --verbose \
    -H "Content-Type: application/json"\
    -H "Accept: application/json" \
    -d '{
        "Username": "johndoe",
        "Password": "12345"
    }'

curl http://localhost:3000/checkjwt --verbose \
    -H "Content-Type: application/json"\
    -H "Authorization: Bearer ${access_token}"

