# Accessing Red Hat Image Repositories
Context: `brew.registry.redhat.io` and `registry.ci.openshift.org` hold Red Hat certified images and you’ll need credentials to pull from these private registries, allowing you building your own images in your local.

## Getting Access to brew.registry.redhat.io
This is a private Red Hat registry. Only authorized Red Hat users will be able to access this image brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_8_golang_1.23 used in the Dockerfile.perses-operator.

Replace <INSERT-USERNAME> with your red hat username. 

```
kinit <INSERT-USERNAME>@IPA.REDHAT.COM
curl --negotiate -k --verbose -u <INSERT-USERNAME> : -X POST -H 'Content-Type: application/json' --data '{"description":"openshift 4 testing"}' https://employee-token-manager.registry.redhat.com/v1/tokens -s | jq 
```

Login to the registry and Copy/Paste username and password from the curl command above when prompted. 
```
podman login brew.registry.redhat.io
```

Reference Doc: https://source.redhat.com/groups/public/teamnado/wiki/brew_registry

## Getting credentials for images from registry.ci.openshift.org
1. Log into https://console-openshift-console.apps.ci.l2s4.p1.openshiftapps.com/ using RH internal SSO
2. User dropdown > Copy login command > copy the oc login command from this page
3. In your terminal, paste the login command from step 2
4. Execute the command oc registry login
5. Now you should be able to pull from that registry <br>
5.1 Output should be something like this:
```
“info: Using registry public hostname registry.ci.openshift.org
Saved credentials for registry.ci.openshift.org into /Users/jezhu/.config/containers/auth.json”
```
(optional) If you have your pull secret for podman set up differently than default, you can use `oc registry login --to <path-to-pull-secret>` to save your creds to whatever pull secret file you use
