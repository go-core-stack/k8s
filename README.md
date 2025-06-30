# k8s
Kubernetes funcs and libraries

## Apply
This pkg is a copy as-is from https://github.com/atsgen/tf-operator/blob/master/pkg/apply

responsible for providing kubectl apply equivalent constructs to be done from a golang based controller

## Render
This pkg is a copy as-is from https://github.com/openshift/cluster-network-operator/tree/master/pkg/render

Render provides go templating engine to render .json or .yaml files based on templates available for k8s objects

The aim is to mimic the parsing behavior of kubectl create -f <dir> as much as reasonably possible.
