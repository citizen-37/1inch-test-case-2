abi-clean:
	rm -rf ./internal/contracts && mkdir ./internal/contracts
abi: abi-clean
	@abigen --abi=./internal/abi/pair.abi --pkg=contracts --type=Pair --out=./internal/contracts/pair.go

mocks-clean:
	rm -rf ./mocks && mkdir ./mocks

mocks: mocks-clean
	@mockery --all --keeptree