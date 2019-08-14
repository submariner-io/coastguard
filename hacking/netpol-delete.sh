#!/bin/bash

kubectl --context=cluster3 delete NetworkPolicy -l app=coastguard

