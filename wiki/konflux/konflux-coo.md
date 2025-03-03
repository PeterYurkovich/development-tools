### Repository
https://github.com/rhobs/konflux-coo

### Add a submodule 
1. Add perses-operator to [.gitmodule](https://github.com/rhobs/konflux-coo/blob/main/.gitmodules). This allows you to point a specific branch in the submodule. 
```
[submodule "perses-operator"]
	path = perses-operator
	url = https://github.com/perses/perses-operator
	branch = release-coo-1.1
```
2. run `git submodule add https://github.com/perses/perses-operator`. This will create the submodule 
3. Add Dockerfile.perses-operator. This is for Konflux to build the image (e.g. perse-operator image). 
This Dockerfile must use Red Hat certified based-images found [here](https://catalog.redhat.com/).
