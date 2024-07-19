# MultiRotate Provider

MultiRotate is a provider that allows simple rotation of multiple versions of a resource.
This is particularly helpful for tokens or service account keys where terraform deletion of a resource immediately invalidates the token/key, and some systems aren't set up to immediately use the new key.
Particularly in the case of things like External Secrets Operator, where depending on the configuration, it can take a long time (relative to the terraform process) for secrets to propagate.
Using "multirotate_set" you can keep 2 (or more) tokens/keys, but always point at the newest version, allowing systems to update to the new token/key over time, rather then having to rearchitect everything to immediately load the new credentials.
