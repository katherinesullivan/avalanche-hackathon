// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import {IRouterClient} from "@chainlink/contracts-ccip/src/v0.8/ccip/interfaces/IRouterClient.sol";
import {Client} from "@chainlink/contracts-ccip/src/v0.8/ccip/libraries/Client.sol";
import {IERC20} from "@chainlink/contracts-ccip/src/v0.8/vendor/openzeppelin-solidity/v4.8.3/contracts/token/ERC20/IERC20.sol";

contract Messenger {
    IERC20 private s_linkToken1;
    IRouterClient private s_router1;


    constructor(address _router1, address _link1) {
        s_linkToken1 = IERC20(_link1);
        s_router1 = IRouterClient(_router1);
    }

    function send(
        uint64 _destinationChainSelector1,
        uint64 _destinationChainSelector2,
        uint64 _destinationChainSelector3,
        uint64 _destinationChainSelector4,
        address _receiver1,
        address _receiver2,
        address _receiver3,
        address _receiver4,
        string calldata _text
    ) external {
        Client.EVM2AnyMessage memory evm2AnyMessage1 = Client.EVM2AnyMessage({
            receiver: abi.encode(_receiver1),
            data: abi.encode(_text),
            tokenAmounts: new Client.EVMTokenAmount[](0),
            extraArgs: "",
            feeToken: address(s_linkToken1)
        });
        Client.EVM2AnyMessage memory evm2AnyMessage2 = Client.EVM2AnyMessage({
            receiver: abi.encode(_receiver2),
            data: abi.encode(_text),
            tokenAmounts: new Client.EVMTokenAmount[](0),
            extraArgs: "",
            feeToken: address(s_linkToken1)
        });
        Client.EVM2AnyMessage memory evm2AnyMessage3 = Client.EVM2AnyMessage({
            receiver: abi.encode(_receiver3),
            data: abi.encode(_text),
            tokenAmounts: new Client.EVMTokenAmount[](0),
            extraArgs: "",
            feeToken: address(s_linkToken1)
        });
        Client.EVM2AnyMessage memory evm2AnyMessage4 = Client.EVM2AnyMessage({
            receiver: abi.encode(_receiver4),
            data: abi.encode(_text),
            tokenAmounts: new Client.EVMTokenAmount[](0),
            extraArgs: "",
            feeToken: address(s_linkToken1)
        });
        
        s_linkToken1.approve(address(s_router1), 1e18 ether);
        s_router1.ccipSend(_destinationChainSelector1, evm2AnyMessage1);

        s_linkToken1.approve(address(s_router1), 1e18 ether);
        s_router1.ccipSend(_destinationChainSelector2, evm2AnyMessage2);

        s_linkToken1.approve(address(s_router1), 1e18 ether);
        s_router1.ccipSend(_destinationChainSelector3, evm2AnyMessage3);

        s_linkToken1.approve(address(s_router1), 1e18 ether);
        s_router1.ccipSend(_destinationChainSelector4, evm2AnyMessage4);

    }
}