	// If the FundMax flag is set, ensure that the acceptable minimum local
	// amount adheres to the amount to be pushed to the remote, and to
	// current rules, while also respecting the protocol-level maximum
	// channel size.
	var minFundAmt, fundUpToMaxAmt chainutil.Amount
	if in.FundMax {
		// Use the protocol-level maximum as the upper bound for our
		// funding attempt.
		if wumboEnabled {
			fundUpToMaxAmt = funding.MaxFLCFundingAmountWumbo
		} else {
			fundUpToMaxAmt = MaxFundingAmount
		}

		// Since the standard non-fundmax flow requires the minimum
		// funding amount to be at least in the amount of the initial
		// remote balance(push amount) we need to adjust the minimum
		// funding amount accordingly. We initially assume the minimum
		// allowed channel size as minimum funding amount.
		minFundAmt = funding.MinChanFundingSize

		// If minFundAmt is less than the initial remote balance we
		// simply assign the initial remote balance to minFundAmt in
		// order to fullfil the criterion. Whether or not this so
		// determined minimum amount is actually available is
		// ascertained downstream in the lnwallet's reservation
		// workflow.
		if remoteInitialBalance >= minFundAmt {
			minFundAmt = remoteInitialBalance
		}
	}
