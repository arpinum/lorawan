package band

import (
	"errors"
	"fmt"
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAU915Band(t *testing.T) {
	Convey("Given the AU 915-928 band is selected", t, func() {
		band, err := GetConfig(AU_915_928, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("When testing the uplink channels", func() {
			testTable := []struct {
				Channel   int
				Frequency int
				DataRates []int
			}{
				{Channel: 0, Frequency: 915200000, DataRates: []int{0, 1, 2, 3, 4, 5}},
				{Channel: 63, Frequency: 927800000, DataRates: []int{0, 1, 2, 3, 4, 5}},
				{Channel: 64, Frequency: 915900000, DataRates: []int{6}},
				{Channel: 71, Frequency: 927100000, DataRates: []int{6}},
			}

			for _, test := range testTable {
				Convey(fmt.Sprintf("Then channel %d must have frequency %d and data rates %v", test.Channel, test.Frequency, test.DataRates), func() {
					So(band.UplinkChannels[test.Channel].Frequency, ShouldEqual, test.Frequency)
					So(band.UplinkChannels[test.Channel].DataRates, ShouldResemble, test.DataRates)
				})
			}
		})

		Convey("When testing the downlink channels", func() {
			testTable := []struct {
				Frequency    int
				DataRate     int
				Channel      int
				ExpFrequency int
			}{
				{Frequency: 915900000, DataRate: 4, Channel: 64, ExpFrequency: 923300000},
				{Frequency: 915200000, DataRate: 3, Channel: 0, ExpFrequency: 923300000},
			}

			for _, test := range testTable {
				Convey(fmt.Sprintf("Then frequency: %d must return frequency: %d", test.Frequency, test.ExpFrequency), func() {
					txChan, err := band.GetUplinkChannelNumber(test.Frequency)
					So(err, ShouldBeNil)
					So(txChan, ShouldEqual, test.Channel)

					freq, err := band.GetRX1Frequency(test.Frequency)
					So(err, ShouldBeNil)
					So(freq, ShouldEqual, test.ExpFrequency)
				})
			}
		})

		Convey("When iterating over all data rates", func() {
			notImplemented := DataRate{}
			for i, d := range band.DataRates {
				if d == notImplemented {
					continue
				}

				expected := i
				if i == 12 {
					expected = 6
				}

				Convey(fmt.Sprintf("Then %v should be DR%d (test %d)", d, expected, i), func() {
					dr, err := band.GetDataRate(d)
					So(err, ShouldBeNil)
					So(dr, ShouldEqual, expected)
				})
			}
		})

		Convey("When testing GetRX1DataRateForOffset", func() {
			testTable := []struct {
				DR       int
				DROffset int
				RX1DR    int
				Error    error
			}{
				{0, 0, 8, nil},
				{0, 1, 8, nil},
				{0, 7, 0, errors.New("lorawan/band: invalid data-rate offset")},
				{4, 0, 12, nil},
				{7, 0, 0, errors.New("lorawan/band: invalid data-rate")},
			}

			for _, test := range testTable {
				Convey(fmt.Sprintf("Given DR %d, DR offset %d", test.DR, test.DROffset), func() {
					dr, err := band.GetRX1DataRate(test.DR, test.DROffset)
					Convey(fmt.Sprintf("Then RX1DR=%d, error=%s", dr, err), func() {
						So(dr, ShouldEqual, test.RX1DR)
						So(err, ShouldResemble, test.Error)
					})
				})
			}
		})

		Convey("When testing GetLinkADRReqPayloadsForEnabledChannels", func() {
			var filteredChans []int
			for i := 8; i < len(band.UplinkChannels); i++ {
				filteredChans = append(filteredChans, i)
			}

			tests := []struct {
				Name                       string
				NodeChannels               []int
				DisableChannels            []int
				EnableChannels             []int
				ExpectedUplinkChannels     []int
				ExpectedLinkADRReqPayloads []lorawan.LinkADRReqPayload
			}{
				{
					Name:                   "all channels active",
					NodeChannels:           band.GetUplinkChannels(),
					ExpectedUplinkChannels: band.GetUplinkChannels(),
				},
				{
					Name:                   "only activate channel 0 - 7",
					NodeChannels:           band.GetEnabledUplinkChannels(),
					DisableChannels:        band.GetEnabledUplinkChannels(),
					EnableChannels:         []int{0, 1, 2, 3, 4, 5, 6, 7},
					ExpectedUplinkChannels: []int{0, 1, 2, 3, 4, 5, 6, 7},
					ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
						{
							Redundancy: lorawan.Redundancy{ChMaskCntl: 7},
						},
						{
							ChMask:     lorawan.ChMask{true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 0},
						},
					},
				},
				{
					Name:                   "only activate channel 8 - 23",
					NodeChannels:           band.GetEnabledUplinkChannels(),
					DisableChannels:        band.GetEnabledUplinkChannels(),
					EnableChannels:         []int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
					ExpectedUplinkChannels: []int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
					ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
						{
							Redundancy: lorawan.Redundancy{ChMaskCntl: 7},
						},
						{
							ChMask:     lorawan.ChMask{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 0},
						},
						{
							ChMask:     lorawan.ChMask{true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 1},
						},
					},
				},
				{
					Name:                   "only activate channel 64 - 71",
					NodeChannels:           band.GetEnabledUplinkChannels(),
					DisableChannels:        band.GetEnabledUplinkChannels(),
					EnableChannels:         []int{64, 65, 66, 67, 68, 69, 70, 71},
					ExpectedUplinkChannels: []int{64, 65, 66, 67, 68, 69, 70, 71},
					ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
						{
							ChMask:     lorawan.ChMask{true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 7},
						},
					},
				},
				{
					Name:                   "only disable channel 0 - 7",
					NodeChannels:           band.GetEnabledUplinkChannels(),
					DisableChannels:        []int{0, 1, 2, 3, 4, 5, 6, 7},
					ExpectedUplinkChannels: filteredChans,
					ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
						{
							ChMask:     lorawan.ChMask{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 0},
						},
					},
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("testing %s [%d]", test.Name, i), func() {
					for _, c := range test.DisableChannels {
						So(band.DisableUplinkChannel(c), ShouldBeNil)
					}
					for _, c := range test.EnableChannels {
						So(band.EnableUplinkChannel(c), ShouldBeNil)
					}
					pls := band.GetLinkADRReqPayloadsForEnabledChannels(test.NodeChannels)
					So(pls, ShouldResemble, test.ExpectedLinkADRReqPayloads)

					chans, err := band.GetEnabledChannelsForLinkADRReqPayloads(test.NodeChannels, pls)
					So(err, ShouldBeNil)
					So(chans, ShouldResemble, test.ExpectedUplinkChannels)
				})
			}
		})
	})
}
