publish:
	ipfs name publish /ipfs/$$(ipfs add -r . | cat | tail -n 1 | awk '{print $$2}')
