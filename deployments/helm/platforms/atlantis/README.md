Github App secrets are encrypted using 
`projects/student-coach-e1e95/locations/asia-southeast1/keyRings/manabie/cryptoKeys/prod-manabie` KMS key
and put into `./templates/secret-gh-app.yaml`.

That secret has 2 fields:
- `key.pem`: the Github App key.
- `webhook-secret`: the webhook secret for Github App.
