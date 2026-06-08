# ExitFleet

ExitFleet is a Go-based, container-native multi-exit proxy gateway inspired by
the VPNGate gateway workflow.

The project is being rebuilt around a manager plus worker model:

- manager: node inventory, scheduling, Docker lifecycle, Web UI, REST API
- worker: one OpenVPN process, one TUN device, one HTTP/SOCKS5 proxy

The original `Guozh1peng/aimili-vpngate` repository is kept separately as a
reference project under `/Users/zhaokang/vpngate/aimili-vpngate`.

