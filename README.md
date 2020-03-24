# Skygaze

Skygaze monitors the Sia network and tries to detect Skynet download activity.

When given a skylink, a Skynet portal will talk to most of the hosts on the Sia
network in an attempt to find the associated skyfile. This project therefore
uses a modified Sia host to listen for incoming sector requests, then tries to
reconstruct the associated skylink and fetches its metadata. The collected
information is provided to the user via a telnet-like server.

To self-host: Patch the Sia source code (see folder `patches`), compile and run
`siad` and configure it as a host and as a Skynet portal. Take note of the
directory that `siad` is running in - let's say it is `~/sia`. In that case run:
`go run main.go ~/sia/skygaze.sock`. Connecting to port 8023 should now provide
access to Skygaze output:

    $ nc localhost 8023
    https://siasky.net/CABAB_1Dt0FJsxqsu_J4TodNCbCGvtFf1Uys_3EgzOlTcg | BigBuckBunny.mp4
