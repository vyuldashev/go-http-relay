# GO HTTP RELAY

Microservice that allows you to carry the proxied traffic over an ordinary HTTP connection.

config.json can be placed inside the executable’s directory or /etc/go-http-relay/.

## Configuration example:
```json
{
  "app_port": "7777",
  "target_url": "https://api.telegram.org",
  "proxy_url": "111.222.333.55:8080",
  "proxy_username": "user1",
  "proxy_password": "pass1"
}
```
- app_port - the port http relay app be listening to;
- target_url - request endpoint;
- proxy_* - proxy settings.
 