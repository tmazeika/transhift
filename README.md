# Transhift

A tiny and simple peer-to-peer file mover.

### Usage

**Local &rightarrow; Remote**

This will continuously try to connect to the remote and then send the file upon connection.

`transhift upload [IP of remote] [password] [path to file]`

**Remote &rightarrow; Local**

This will wait for a connection to arrive from the specified remote IP and then download the file.

`transhift download [password] [IP of remote]`

*Options*

* `--destination -d [path to dest. file]` Saves the remote file to the given destination
