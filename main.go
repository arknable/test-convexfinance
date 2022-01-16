package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	earmarkIncentive, err := booster.EarmarkIncentive(&bind.CallOpts{Context: ctx})
	if err != nil {
		exitWithError(logger, err)
	}

	crvAddr, err := booster.Crv(&bind.CallOpts{Context: ctx})
	if err != nil {
		exitWithError(logger, err)
	}
	logger.Log("CRV Address: ", crvAddr)

	done := false
	intChan := make(chan os.Signal)
	signal.Notify(intChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-intChan
		done = true
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			crvBalance, err := client.BalanceAt(ctx, crvAddr, nil)
			if err != nil {
				exitWithError(logger, err)
			}

			callIncentive := crvBalance.Mul(crvBalance, earmarkIncentive)
			callIncentive = callIncentive.Div(callIncentive, feeDenominator)

			fmt.Println()
			fmt.Println("Fee Denominator: ", feeDenominator)
			fmt.Println("Earmark Incentive: ", earmarkIncentive)
			fmt.Println("CRV Balance: ", crvBalance)
			fmt.Println("Call Incentive is ", callIncentive)

			if done {
				wg.Done()
				break
			}
			time.Sleep(time.Second)
		}
	}()

	wg.Wait()
	fmt.Println("Stopped")
}

func exitWithError(logger log.Logger, err error) {
	logger.Log("ERR: ", err)
	os.Exit(1)
}