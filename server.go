		AuxCloser: fn.MapOption(
			func(c chcl.AuxChanCloser) contractcourt.AuxChanCloser {
				return c
			},
		)(s.implCfg.AuxChanCloser),
		ChannelCloseConfs: cfg.Dev.ChannelCloseConfs(),
	}, dbs.ChanStateDB)
