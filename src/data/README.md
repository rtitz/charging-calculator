# Data dir

This directory should contain the output json files from wallbox-monitor and solar-monitor.

```
rsync -avz --include='*/' --include='*.json' --exclude='*' --update --ignore-existing 192.168.1.4:/root/{solar,wallbox}-monitor/ .
```
