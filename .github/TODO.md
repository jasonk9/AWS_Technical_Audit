# TO DO

### v1.0.0

- [x] Add flags to parse command args and set defaults
- [x] Add check for file exists to get policy funcs
- [x] Update func signature to be more reusable func GetAccountsFromGithub( repo string, owner string, fileName string ) ( []accountInfo, err )
- [x] Change main call to GetAccountsFromGithub to use new signature

### v1.0.1

- [ ] Update GetFromGithub to use the live site
- [ ] Add debug logging
- [ ] Add Console Feedback

### Future

- [ ] Add IAM Identity Center Auditing
- [ ] Add auto fetch of RolePolicy when assumeRole