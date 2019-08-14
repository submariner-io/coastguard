# coastguard
Controller to facilitate network policing on a multi-cluster connected environments (proof-of-concept state)


# setup development environment
You will need docker installed in your system, and at least 8GB of RAM.

This will deploy 3 kind clusters + submariner + coastguard, and run the 
coastguard e2e tests on top.

```bash
   make e2e status=keep
```

This will update the coastguard images and containers, and can be executed
on top of an already existing 3 kind clusters + submariner [+coastguard]
deployment:

```bash
   make e2e-coastguard status=keep
```

# testing

## run e2e testing
```bash
   make e2e
```

## run unit testing
```bash
   make test
```


