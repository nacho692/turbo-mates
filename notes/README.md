#### 21/11/2022

Just implemented some basic lookup/friends response via UDP.
The peer discovery algorithm is based on kademlias ideas, symmetry on  
distance  and "handshaking" by protocol is a good NTH.

I'm not completely convinced that having a tree like k-bucket structure is 
necessary if i'm not looking to implement a distributed hash table.
But the distribution of peers via clusters of distance seems to provide 
good protection against network partitioning.
I am not limiting the size of the buckets for now but will do in the future, 
maybe limiting further buckets more than nearer ones.

Each node sets up a UDP socket to listen and write. 
If configured, they send a *lookup* message to a bootstrap node, looking for 
its own ID which in turn responds with a *friends* message indicating the 
*alpha* closest peers it knows.
If any of the received peers is closer than the previous peer, the requester 
should re-ask a *lookup* to it. 
This requires having some sort of state between messages, as the requester 
doesn't know when it receives the response the previous minimum distance it 
had. 
There are two possible solutions that come to mind:
* Adding a *lookup* ID, adding this ID to the *friends* response and storing 
  in memory the previous distance accessed by the *lookup* ID.
* Allowing the possibility to concatenate some kind of generic payload from 
  the request to the response. Instead of adding the request ID in the 
  response, adding the previous distance itself.

I generally dislike having to track mutability everywhere, and the cache 
tends to go that way. I think I'll be adding an ID to the messages and also 
a generic request to response payload to be able to track state across messages.
This even allows having a history of request/response series to nodes, even 
if it increases the payload size.