panacead tx gov submit-proposal software-upgrade "v2.0.6" --title="test" --description="test" --deposit 10umed --upgrade-height 20 --upgrade-info "2.0.5-12-g93409f6" --from oracle1 --chain-id local --fees 10000000umed

panacead tx gov vote 1 yes --from oracle1 --chain-id local --fees 10000000umed -y
