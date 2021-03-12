package keeper_test

import (
	"github.com/hyperledger/burrow/execution/errors"

	"github.com/certikfoundation/shentu/x/cvm/types"
)

var (
	InsufficientGasErrorCode   = types.BurrowErrorCodeStart + errors.Codes.InsufficientGas.ErrorCode().Number
	CodeOutOfBoundsErrorCode   = types.BurrowErrorCodeStart + errors.Codes.CodeOutOfBounds.ErrorCode().Number
	ExecutionRevertedErrorCode = types.BurrowErrorCodeStart + errors.Codes.ExecutionReverted.ErrorCode().Number
	/*
		pragma solidity >=0.4.22 <0.6.0;
		contract Hello55 {
			function sayHi() public pure returns (uint) {
				return 55;
			}
		}
	*/
	Hello55BytecodeString     = "6080604052348015600f57600080fd5b5060ac8061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c80630c49c36c14602d575b600080fd5b60336049565b6040518082815260200191505060405180910390f35b6000603790509056fea26469706673582212207e355097621bbba6da748b17d355f936e7b5ad809077d16433f46dbfda2cae0364736f6c637826302e362e342d646576656c6f702e323032302e332e352b636f6d6d69742e33326361316135650057"
	Hello55AbiJsonString      = `[{"inputs":[],"name":"sayHi","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"pure","type":"function"}]`
	Hello55MetadataJsonString = `{"compiler":{"version":"0.6.4-develop.2020.3.5+commit.32ca1a5e"},"language":"Solidity","output":{"abi":[{"inputs":[],"name":"sayHi","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"pure","type":"function"}],"devdoc":{"methods":{}},"userdoc":{"methods":{}}},"settings":{"compilationTarget":{"tests/hello55.sol":"Hello55"},"evmVersion":"istanbul","libraries":{},"metadata":{"bytecodeHash":"ipfs"},"optimizer":{"enabled":false,"runs":200},"remappings":[]},"sources":{"tests/hello55.sol":{"keccak256":"0x743a9d971915a1fe43a196d542a13ab0f4f36f9e3c27579eb3e64b78a0469182","urls":["bzz-raw://55a4dce17dbdde893afbfde0cbe9dd4ff7b9a281e1e3e4a47a1ea7a8733aa1de","dweb:/ipfs/QmTR4LxBT2F38RzuCogTjpk8iEXRpz2RRNjeLB8Bwc6SQs"]}},"version":1}`

	/*
		pragma solidity >=0.4.22 <0.6.0;
		contract BasicTests {
			uint myFavoriteNumber = 34;
			function addTwoNumbers(uint a, uint b) public pure returns (uint) {
				return a + b;
			}
			function failureFunction() public pure {
				revert("Go away!!");
			}
			function setMyFavoriteNumber(uint newFavNum) public {
				myFavoriteNumber = newFavNum;
			}
		}
	*/
	BasicTestsBytecodeString = "6080604052602260005534801561001557600080fd5b50610184806100256000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c80630b30d76414610046578063a0eb379f14610074578063e2276f1c1461007e575b600080fd5b6100726004803603602081101561005c57600080fd5b81019080803590602001909291905050506100ca565b005b61007c6100d4565b005b6100b46004803603604081101561009457600080fd5b810190808035906020019092919080359060200190929190505050610142565b6040518082815260200191505060405180910390f35b8060008190555050565b6040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260098152602001807f476f20617761792121000000000000000000000000000000000000000000000081525060200191505060405180910390fd5b600081830190509291505056fea265627a7a7231582029e87152c00d34140b78a06d51e5b41bdd4eab369148d1b9540394dcc93f1d5e64736f6c634300050b0032"
	BasicTestsAbiJsonString  = `
	[
	{
		"constant": false,
		"inputs": [
			{
				"internalType": "uint256",
				"name": "newFavNum",
				"type": "uint256"
			}
		],
		"name": "setMyFavoriteNumber",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "failureFunction",
		"outputs": [],
		"payable": false,
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{
				"internalType": "uint256",
				"name": "a",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "b",
				"type": "uint256"
			}
		],
		"name": "addTwoNumbers",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "pure",
		"type": "function"
	}
]`
	/*
		pragma solidity >=0.4.22 <0.6.0;
		contract GoofyContract {
			uint public goofyNumber;
			function setGoofyNumber(uint a) public {
				goofyNumber = a;
			}
		}
		contract GasTests {
			GoofyContract gc;
			function addTwoNumbers(uint a, uint b) public pure returns (uint) {
				return a + b;
			}
			function hashMe(bytes memory b) public pure returns (bytes32) {
				return keccak256(b);
			}
			function deployAnotherContract() public {
				gc = new GoofyContract();
			}
			function setGoofyNumber(uint a) public returns (uint) {
				gc.setGoofyNumber(a);
				return gc.goofyNumber();
			}
		}
	*/
	GasTestsBytecodeString = "608060405234801561001057600080fd5b5061049a806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80632e60e461146100515780635422426314610120578063b71b449014610162578063e2276f1c1461016c575b600080fd5b61010a6004803603602081101561006757600080fd5b810190808035906020019064010000000081111561008457600080fd5b82018360208201111561009657600080fd5b803590602001918460018302840111640100000000831117156100b857600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506101b8565b6040518082815260200191505060405180910390f35b61014c6004803603602081101561013657600080fd5b81019080803590602001909291905050506101c9565b6040518082815260200191505060405180910390f35b61016a610300565b005b6101a26004803603604081101561018257600080fd5b81019080803590602001909291908035906020019092919050505061036a565b6040518082815260200191505060405180910390f35b600081805190602001209050919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166354224263836040518263ffffffff1660e01b815260040180828152602001915050600060405180830381600087803b15801561023f57600080fd5b505af1158015610253573d6000803e3d6000fd5b505050506000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16630a2346c66040518163ffffffff1660e01b815260040160206040518083038186803b1580156102be57600080fd5b505afa1580156102d2573d6000803e3d6000fd5b505050506040513d60208110156102e857600080fd5b81019080805190602001909291905050509050919050565b60405161030c90610377565b604051809103906000f080158015610328573d6000803e3d6000fd5b506000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550565b6000818301905092915050565b60e2806103848339019056fe608060405234801561001057600080fd5b5060c38061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060325760003560e01c80630a2346c614603757806354224263146053575b600080fd5b603d607e565b6040518082815260200191505060405180910390f35b607c60048036036020811015606757600080fd5b81019080803590602001909291905050506084565b005b60005481565b806000819055505056fea265627a7a72315820bd31260ca27a654607a2d4b452d03506665f2e48c15e43ae924552f645d26de864736f6c634300050c0032a265627a7a723158209366e5562db198cc3463c52ac5b103071ee3b30d645e47b876037d90f7c4b6d564736f6c634300050c0032"
	GasTestsAbiJsonString  = `
	[
		{
			"constant": true,
			"inputs": [
				{
					"internalType": "bytes",
					"name": "b",
					"type": "bytes"
				}
			],
			"name": "hashMe",
			"outputs": [
				{
					"internalType": "bytes32",
					"name": "",
					"type": "bytes32"
				}
			],
			"payable": false,
			"stateMutability": "pure",
			"type": "function"
		},
		{
			"constant": false,
			"inputs": [
				{
					"internalType": "uint256",
					"name": "a",
					"type": "uint256"
				}
			],
			"name": "setGoofyNumber",
			"outputs": [
				{
					"internalType": "uint256",
					"name": "",
					"type": "uint256"
				}
			],
			"payable": false,
			"stateMutability": "nonpayable",
			"type": "function"
		},
		{
			"constant": false,
			"inputs": [],
			"name": "deployAnotherContract",
			"outputs": [],
			"payable": false,
			"stateMutability": "nonpayable",
			"type": "function"
		},
		{
			"constant": true,
			"inputs": [
				{
					"internalType": "uint256",
					"name": "a",
					"type": "uint256"
				},
				{
					"internalType": "uint256",
					"name": "b",
					"type": "uint256"
				}
			],
			"name": "addTwoNumbers",
			"outputs": [
				{
					"internalType": "uint256",
					"name": "",
					"type": "uint256"
				}
			],
			"payable": false,
			"stateMutability": "pure",
			"type": "function"
		}
	]
	`
	//derived using remix.ethereum.org
	AddTwoNumbersGasCost         uint64 = 13055
	HashMeGasCost                uint64 = 13099
	DeployAnotherContractGasCost uint64 = 113549 + 92277
	SetGoofyNumberGasCost        uint64 = 47895 + 26367

	/*
		pragma solidity >=0.4.22 <0.6.0;
		contract GasRefund {
			uint stupidNumber;
			constructor() public {
				stupidNumber=10000000;
			}
			function iWillRevert() public {
				uint a = 4 + 5;
				revert("thats enough work for now");
			}
			function iWillFail() public {
				uint a = 4 - 4;
				uint b = 6 / a;
			}
			function deleteFromStorage() public {
				stupidNumber = 0;
			}
			function die() public {
				selfdestruct(address(0x0));
			}
		}
	*/
	//GasRefundBytecodeString = /* version with self destruct address as longer string */ "608060405234801561001057600080fd5b50629896806000819055506101708061002a6000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c806335f469941461005157806360a5f1871461005b57806383bfecfa146100655780639e2ff0741461006f575b600080fd5b610059610079565b005b6100636100a6565b005b61006d6100af565b005b6100776100c7565b005b73ab35ee8df2f8dd950cc1cfd38fef86857374e97173ffffffffffffffffffffffffffffffffffffffff16ff5b60008081905550565b60008090506000816006816100c057fe5b0490505050565b6000600990506040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260198152602001807f746861747320656e6f75676820776f726b20666f72206e6f770000000000000081525060200191505060405180910390fdfea265627a7a72315820e6b239ca4c62450752d8bb2877fe37bc15f1cc95afb6f9387df4199c97f6b6c864736f6c634300050c0032"
	GasRefundBytecodeString = /* version with self destruct address as zero */ "608060405234801561001057600080fd5b506298968060008190555061015d8061002a6000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c806335f469941461005157806360a5f1871461005b57806383bfecfa146100655780639e2ff0741461006f575b600080fd5b610059610079565b005b610063610093565b005b61006d61009c565b005b6100776100b4565b005b600073ffffffffffffffffffffffffffffffffffffffff16ff5b60008081905550565b60008090506000816006816100ad57fe5b0490505050565b6000600990506040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260198152602001807f746861747320656e6f75676820776f726b20666f72206e6f770000000000000081525060200191505060405180910390fdfea265627a7a7231582086b971c2ec8ef8b8fb46a876f47f527b6e879ecd64bb16d3d864febac60b29c364736f6c634300050c0032"
	GasRefundAbiJsonString  = `
		[
			{
				"inputs": [],
				"payable": false,
				"stateMutability": "nonpayable",
				"type": "constructor"
			},
			{
				"constant": false,
				"inputs": [],
				"name": "deleteFromStorage",
				"outputs": [],
				"payable": false,
				"stateMutability": "nonpayable",
				"type": "function"
			},
			{
				"constant": false,
				"inputs": [],
				"name": "die",
				"outputs": [],
				"payable": false,
				"stateMutability": "nonpayable",
				"type": "function"
			},
			{
				"constant": false,
				"inputs": [],
				"name": "iWillFail",
				"outputs": [],
				"payable": false,
				"stateMutability": "nonpayable",
				"type": "function"
			},
			{
				"constant": false,
				"inputs": [],
				"name": "iWillRevert",
				"outputs": [],
				"payable": false,
				"stateMutability": "nonpayable",
				"type": "function"
			}
		]
		`

	CtkTransferTestBytecodeString = "608060405234801561001057600080fd5b50610145806100206000396000f3fe6080604052600436106100295760003560e01c80634897e0631461002e578063eb53b14e14610072575b600080fd5b6100706004803603602081101561004457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061009d565b005b34801561007e57600080fd5b506100876100f1565b6040518082815260200191505060405180910390f35b8073ffffffffffffffffffffffffffffffffffffffff166108fc600234816100c157fe5b049081150290604051600060405180830381858888f193505050501580156100ed573d6000803e3d6000fd5b5050565b60003073ffffffffffffffffffffffffffffffffffffffff163190509056fea265627a7a72315820995e9941f88fede53e75c0509196b74084b2cb6778b668a6cddee82c542e075564736f6c634300050c0032"
	CtkTransferTestAbiJsonString  = `
			[
				{
					"constant": false,
					"inputs": [
						{
							"internalType": "address payable",
							"name": "friend",
							"type": "address"
						}
					],
					"name": "sendToAFriend",
					"outputs": [],
					"payable": true,
					"stateMutability": "payable",
					"type": "function"
				},
				{
					"constant": true,
					"inputs": [],
					"name": "whatsMyBalance",
					"outputs": [
						{
							"internalType": "uint256",
							"name": "",
							"type": "uint256"
						}
					],
					"payable": false,
					"stateMutability": "view",
					"type": "function"
				}
			]
			`

	TestCheckBytecodeString = "60806040526040518060600160405280602d8152602001610812602d91396000908051906020019061003292919061014d565b506040518060600160405280602d815260200161086c602d91396001908051906020019061006192919061014d565b506040518060600160405280602d815260200161083f602d91396002908051906020019061009092919061014d565b506040518060600160405280602d8152602001610812602d9139600390805190602001906100bf92919061014d565b506040518060600160405280602d815260200161086c602d9139600490805190602001906100ee92919061014d565b506040518060400160405280601381526020017f64756d6d79736f75726365636f646568617368000000000000000000000000008152506005908051906020019061013a92919061014d565b5034801561014757600080fd5b50610251565b828054610159906101f0565b90600052602060002090601f01602090048101928261017b57600085556101c2565b82601f1061019457805160ff19168380011785556101c2565b828001600101855582156101c2579182015b828111156101c15782518255916020019190600101906101a6565b5b5090506101cf91906101d3565b5090565b5b808211156101ec5760008160009055506001016101d4565b5090565b6000600282049050600182168061020857607f821691505b6020821081141561021c5761021b610222565b5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b6105b2806102606000396000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c80630bfe74e81461005c5780634a6abcda1461007a5780638504647c146100985780639ca131af146100b6578063de78f9fc146100d4575b600080fd5b6100646100f2565b6040516100719190610499565b60405180910390f35b6100826101c1565b60405161008f9190610499565b60405180910390f35b6100a0610269565b6040516100ad9190610499565b60405180910390f35b6100be610310565b6040516100cb9190610499565b60405180910390f35b6100dc6103b8565b6040516100e99190610499565b60405180910390f35b60606000600480546101039061050a565b80601f016020809104026020016040519081016040528092919081815260200182805461012f9061050a565b801561017c5780601f106101515761010080835404028352916020019161017c565b820191906000526020600020905b81548152906001019060200180831161015f57829003601f168201915b50505050509050600181516001828260208601606561c350fa6001838360208701606661c350fa60018214156101bb5760018114156101ba57600180f35b5b60016000f35b60606000600180546101d29061050a565b80601f01602080910402602001604051908101604052809291908181526020018280546101fe9061050a565b801561024b5780601f106102205761010080835404028352916020019161024b565b820191906000526020600020905b81548152906001019060200180831161022e57829003601f168201915b50505050509050600181516001828260208601606561c350fa600183f35b606060008080546102799061050a565b80601f01602080910402602001604051908101604052809291908181526020018280546102a59061050a565b80156102f25780601f106102c7576101008083540402835291602001916102f2565b820191906000526020600020905b8154815290600101906020018083116102d557829003601f168201915b50505050509050600181516001828260208601606561c350fa600183f35b60606000600280546103219061050a565b80601f016020809104026020016040519081016040528092919081815260200182805461034d9061050a565b801561039a5780601f1061036f5761010080835404028352916020019161039a565b820191906000526020600020905b81548152906001019060200180831161037d57829003601f168201915b50505050509050600181516001828260208601606661c350fa600183f35b60606000600580546103c99061050a565b80601f01602080910402602001604051908101604052809291908181526020018280546103f59061050a565b80156104425780601f1061041757610100808354040283529160200191610442565b820191906000526020600020905b81548152906001019060200180831161042557829003601f168201915b50505050509050600181516001828260208601606761c350fa600183f35b600061046b826104bb565b61047581856104c6565b93506104858185602086016104d7565b61048e8161056b565b840191505092915050565b600060208201905081810360008301526104b38184610460565b905092915050565b600081519050919050565b600082825260208201905092915050565b60005b838110156104f55780820151818401526020810190506104da565b83811115610504576000848401525b50505050565b6000600282049050600182168061052257607f821691505b602082108114156105365761053561053c565b5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b6000601f19601f830116905091905056fea2646970667358221220169b18dba603a9571f2f7778926c03823b3affd9b07cb7f74781bc87379ab4c864736f6c6343000801003363657274696b3175647a7032337477663461367066343765723236756a67637a7a683374637374376335617a65636f736d6f7331723630686a327861786e373971746834706b6a6d397432376c3938357866736d6e7a39706177636f736d6f733178786b75656b6c616c3976656a7639756e717538307739767074796570666139357064353375"
	TestCheckAbiJsonString  = `[{"inputs":[],"name":"callCheck","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"callCheckNotCertified","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"compilationCheck","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"proofAndAuditingCheck","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"proofCheck","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"nonpayable","type":"function"}]`

	TestCertifyValidatorString        = "60806040526040518060800160405280605381526020016103a16053913960009080519060200190610032929190610045565b5034801561003f57600080fd5b50610149565b828054610051906100e8565b90600052602060002090601f01602090048101928261007357600085556100ba565b82601f1061008c57805160ff19168380011785556100ba565b828001600101855582156100ba579182015b828111156100b957825182559160200191906001019061009e565b5b5090506100c791906100cb565b5090565b5b808211156100e45760008160009055506001016100cc565b5090565b6000600282049050600182168061010057607f821691505b602082108114156101145761011361011a565b5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b610249806101586000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80633c1bf57b14610030575b600080fd5b61003861004e565b6040516100459190610130565b60405180910390f35b6060600080805461005e906101a1565b80601f016020809104026020016040519081016040528092919081815260200182805461008a906101a1565b80156100d75780601f106100ac576101008083540402835291602001916100d7565b820191906000526020600020905b8154815290600101906020018083116100ba57829003601f168201915b505050505090506001815160018282602086016000606861c350f1600183f35b600061010282610152565b61010c818561015d565b935061011c81856020860161016e565b61012581610202565b840191505092915050565b6000602082019050818103600083015261014a81846100f7565b905092915050565b600081519050919050565b600082825260208201905092915050565b60005b8381101561018c578082015181840152602081019050610171565b8381111561019b576000848401525b50505050565b600060028204905060018216806101b957607f821691505b602082108114156101cd576101cc6101d3565b5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b6000601f19601f830116905091905056fea26469706673582212206432f04b3863a71e348225305d55be91def584a696225e327741a5583432d26764736f6c63430008010033636f736d6f7376616c636f6e73707562317a636a647565707178687936383635686639306c776d636b6a756567666476716d797a6e6864366134646b6a72393070713061383266787867327171637066716174"
	TestCertifyValidatorAbiJsonString = `[{"inputs":[],"name":"certifyValidator","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"nonpayable","type":"function"}]`
)