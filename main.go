package main

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/cryptoriums/packages/ethereum"
	"github.com/cryptoriums/packages/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-kit/log"

	"github.com/arknable/test-convexfinance/contracts"
)

/*
	Call incentive is calculated in _earmarkRewards() as follows,

	__________
	function _earmarkRewards(uint256 _pid) internal {
		.....
		.....

		//crv balance
		uint256 crvBal = IERC20(crv).balanceOf(address(this));

		if (crvBal > 0) {
			uint256 _lockIncentive = crvBal.mul(lockIncentive).div(FEE_DENOMINATOR);
			uint256 _stakerIncentive = crvBal.mul(stakerIncentive).div(FEE_DENOMINATOR);
			uint256 _callIncentive = crvBal.mul(earmarkIncentive).div(FEE_DENOMINATOR);

			.....
			.....

			//send incentives for calling
			IERC20(crv).safeTransfer(msg.sender, _callIncentive);

			.....
			.....
		}
	}
	__________

	The only variable is `earmarkIncentive` so we can calculate the incentive as is using the formula above.
*/

const BoosterContractAddress = "0xf403c135812408bfbe8713b5a23a04b3d48aae31"

func main() {
	nodeURL := os.Getenv("NODE_URLS")
	if nodeURL == "" {
		fmt.Println("Please set NODE_URLS environment variable.")
		os.Exit(0)
	}

	logger := logging.NewLogger()
	ctx := context.Background()
	client, err := ethereum.NewClient(ctx, logger, map[string]string{
		"NODE_URLS": nodeURL,
	})
	if err != nil {
		exitWithError(logger, err)
	}
	defer client.Close()

	booster, err := contracts.NewBooster(common.HexToAddress(BoosterContractAddress), client)
	if err != nil {
		exitWithError(logger, err)
	}

	feeDenominator, err := booster.FEEDENOMINATOR(&bind.CallOpts{Context: ctx})
	if err != nil {
		exitWithError(logger, err)
	}
	logger.Log("Fee Denominator: ", feeDenominator)

	earmarkIncentive, err := booster.EarmarkIncentive(&bind.CallOpts{Context: ctx})
	if err != nil {
		exitWithError(logger, err)
	}
	logger.Log("Earmark Incentive: ", earmarkIncentive)

	crvAddr, err := booster.Crv(&bind.CallOpts{Context: ctx})
	if err != nil {
		exitWithError(logger, err)
	}
	logger.Log("CRV Address: ", crvAddr)

	// Original code:
	// uint256 crvBal = IERC20(crv).balanceOf(address(this));
	//
	// so we can assume a value as current balance
	// because crvBal is balance of the wallet which connected to the contract.
	balance := big.NewInt(337)
	logger.Log("CRV Balance: ", balance)

	callIncentive := balance.Mul(balance, earmarkIncentive)
	callIncentive = callIncentive.Div(callIncentive, feeDenominator)

	fmt.Println("Call Incentive: ", callIncentive)
}

func exitWithError(logger log.Logger, err error) {
	logger.Log("ERR: ", err)
	os.Exit(1)
}