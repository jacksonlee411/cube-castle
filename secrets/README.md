# Secrets Directory (Do Not Commit Secrets)

- This folder stores local-only development secrets (e.g., RS256 keypair).
- Keys must never be committed to the repository. They are intentionally ignored via `.gitignore`.
- For development, generate keys on demand:

```
make jwt-dev-setup
```

- This will create `secrets/dev-jwt-private.pem` and `secrets/dev-jwt-public.pem` locally if absent.
- If you need to rotate keys, delete the existing files and re-run the command.

Security reminders:
- Do not share keys or tokens in logs or commits.
- If keys were accidentally committed in the past, rotate them and remove from history (BFG or filter-repo).

