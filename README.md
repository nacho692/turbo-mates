# turbo-mates

Turbo Mates is a... well, it wants to be a distributed append only log akin 
to blockchain but without security concerns.

I'll be trying to divide the development in three main areas.

* Peer Discovery: Keeps an updated table of working peers.
* Visualization: It would be awesome to understand and explore the network 
  via some visualization tools, maybe exposing an interface to hook your own 
  graphical display.
* Log distribution: Handles new node synchronization and broadcasting of new 
  logs.

It would be nice to mount some kind of application over this infrastructure 
protocol such as a small game or a database.

I'm also writing some notes on development desitions and issues I'm facing on 
[notes/README.md](notes/README.md).