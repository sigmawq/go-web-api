# API Server 

# Installation
`go build src/.`

# Usage
1. Launch the executable and the server is up.
2. Click enter to close it.
3. Use anything to connect to it via HTTP.

Note: page query syntax is slightly complex:
`/api/v1/users?page=<value>&order_by=<[id | firstname | secondname | age | lon | lat]>&filter_by=<[id | firstname | secondname | age | lon | lat]>
// .. &filter_pred=[g | l | ge | le | e | ne]&filter_value=<string>`

For example: All users with an age value of less than or equal to 40, ordered by id:
`/api/v1/users?page=0&order_by=id&filter_by=age&filter_pred=le&filter_value=40`