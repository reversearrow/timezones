# timezones
Go Microservice  to Query Time zones 

## Usage

```
On localhost:

➜ curl http://localhost:8081/api/time\?tz=America/New_York,Asia/Kolkata
{
  "America/New_York": "2022-12-26 16:12:30.742344782 -0500 EST",
  "Asia/Kolkata": "2022-12-27 02:42:30.742344782 +0530 IST"
}
```