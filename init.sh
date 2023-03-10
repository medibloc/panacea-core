rm -rf ~/.panacea
panacead init local --chain-id local

panacead keys add oracle1
panacead add-genesis-account $(panacead keys show oracle1 -a) 100000000000umed
panacead gentx oracle1 100000000umed --commission-rate 0.1 --commission-max-rate 0.2 --commission-max-change-rate 0.01  --min-self-delegation 1 --chain-id local

panacead collect-gentxs
