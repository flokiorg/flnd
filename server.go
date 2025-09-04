		NumRequiredConfs: func(chanAmt chainutil.Amount,
			pushAmt lnwire.MilliLoki) uint16 {

			return lnwallet.FundingConfsForAmounts(
				chanAmt, pushAmt,
			)
		},
