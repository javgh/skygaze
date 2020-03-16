# Skygaze

Skygaze monitors the Sia network and tries to detect Skynet download activity.

When given a skylink, a Skynet portal will talk to most of the hosts on the Sia
network in an attempt to find the associated skyfile. This project therefore
uses a modified Sia host to listen for incoming sector requests, then tries to
reconstruct the associated skylink and fetches its metadata. The collected
information is provided to the user via a telnet-like server.

One instance of such a server is available as a Tor hidden service at
`l2u45pvrhpnyjdw2g2ulenydp3c3jiywntjiktzqp6foqhy4wzipfvyd.onion:23`.
Connect using netcat:

    $ torify nc l2u45pvrhpnyjdw2g2ulenydp3c3jiywntjiktzqp6foqhy4wzipfvyd.onion 23
    https://siasky.net/CABAB_1Dt0FJsxqsu_J4TodNCbCGvtFf1Uys_3EgzOlTcg | BigBuckBunny.mp4

To self-host: Patch the Sia source code (see folder `patches`), compile and run
`siad`. Take note of the directory that `siad` is running in - let's say it is
`~/sia`. In that case run: `go run main.go ~/sia/skygaze.sock`. Connecting to
port 8023 should now provide access to Skygaze output: `nc localhost 8023`.
