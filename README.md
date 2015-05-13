# InstantRoom

InstantRoom is a simple, private and secure group chat application.

Although they are many existing group chat application, InstantRoom is different in that
it avoids complexity by putting back the responsibility of the security back
into the user's hands.

Indeed it is the responsibility of group members to initially transmit the group secret key to other members.

InstantRoom makes that easy with:
- _a secret key with mini format for easy sharing_
- _or a secret key as QR code for direct phone to phone sharing_

InstantRoom does not persist any messages on the server nor can it decrypt any messages since the secret key is only
generated and managed on the user's client application side.

InstantRoom abide by the Kerckhoffs's principle:

> A cryptosystem should be secure even if everything about the system, except the key, is public knowledge

## Features

- Secret keys have a mini format for easy sharing
- Secret keys can be displayed as a QR code for phone to phone sharing
- Rooms auto expire. One can choose a 1 hour room, 1 day room, 1 week room, etc...
- Messages are only local on the client side. You can chooses to delete some or all of them at anytime
- The application's code is in the open and inspectable by anyone
- The server does not store any messages
- The server cannot decrypt any messages. It only relays them

## Upcoming features

1. Local access to secret keys on the client side will be password protected
2. Your own InstantRoom! Executables will be available to install your own InstantRoom server. The client applications will allow to point to any server's url

## Attack scenarios

InstantRoom is based on the secrecy of your private key generated through the Public Key Infrastructure. So leaving aside advanced implementation attacks scenario, here is the only way your group chat can be compromised:

- the attacker gets hold of your secret key!

How?

- if you leave your phone unattended (although in the future the access of the secret key on your phone will be password protected)
- if you have transmitted the secret key to other members through a potentially insecure channel

## Sharing the secret key

Here would be common way to give the secret key to other members and the associated risk:

Method | Attack | Risk
--- | --- | ---
Orally | being heard by third party | unlikely (as can be easily circumvented)
Direct contact reading through QR code |  side channel | unlikely (as very elaborate and spottable)
Remote webcam reading through QR code | insecure channel | possible

## Kerckhoffs's six design principles

We believe InstantRoom stands by those principles

Principles| Status
---|---
The system must be practically, if not mathematically, indecipherable|&#10003;
It should not require secrecy, and it should not be a problem if it falls into enemy hands|&#10003;
It must be possible to communicate and remember the key without using written notes, and correspondents must be able to change or modify it at will|&#10003;
It must be applicable to telegraph communications|&#10003;
It must be portable, and should not require several persons to handle or operate|&#10003;
The system must be easy to use and should not be stressful to use or require its users to know and comply with a long list of rules|&#10003;


##  Technicalities

- Private and public keys are generated through ECDSA P-256
- The mini private key format is done through the Bitcoin's 30 characters format
- The encryption is done through AES using the private key as the cypher key. The key size is therefore 256 bits.


