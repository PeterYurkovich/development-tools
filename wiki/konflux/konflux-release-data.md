## Cloning 
Cloning https://gitlab.cee.redhat.com/releng/konflux-release-data requires authentication with a token (and VPN). Here are the instructions below on how to git clone with authentication token:
1. Create a personal access token https://docs.gitlab.com/user/profile/personal_access_tokens/
2. git clone  `https://<username>:<TOKEN>@github.com/<USERNAME>/<REPO>.git` (e.g. git clone https://jezhu:abc123abc123abc123@gitlab.cee.redhat.com/jezhu/konflux-release-data.git)
## Running tests on local machine  
1. Install `tox` and `yp` (brew install tox, brew install yp)
2. Run `tox` to run test suite
