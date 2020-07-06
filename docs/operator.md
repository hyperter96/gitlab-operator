# Introduction
The operator pattern provides a way to package, deploy and manager Kubernetes applications.

Operators often extend the Kubernetes API by introducing a new resource types using Kubernetes CRDs and uses domain specific information to manage the said resource types.

## Gitlab Operator
The Gitlab Operator introduces two new resource types at this time:  
* [Gitlab](gitlab.md)
* [Runner](runner.md)

The Gitlab resource would deploy the Gitlab application and or the registry depending on the resource spec provided when creating the resource.

The Runner resource would be used to create a resource that can be used to integrate with the
