package account_test

import (
	"strings"
	"time"

	"github.com/iotaledger/iota.go/account"
	"github.com/iotaledger/iota.go/account/builder"
	"github.com/iotaledger/iota.go/account/deposit"
	"github.com/iotaledger/iota.go/account/event"
	"github.com/iotaledger/iota.go/account/event/listener"
	"github.com/iotaledger/iota.go/account/plugins/transfer/poller"
	"github.com/iotaledger/iota.go/account/store/inmemory"
	. "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

const seed = "BXY9DNFAJKY9UPRTP9AQFQKESGLBDXBQHIXSFDDBLIJDXSKLBYFJ9XPZ9YSAGFVXQHBODMNFLNCMAJHYZ"
const id = "d7e75aa9def2ef9c813313f0e0fb72b9"
const usedSecLvl = consts.SecurityLevelLow

type fakeclock struct{}

func (fc *fakeclock) Time() (time.Time, error) {
	return time.Date(2018, time.November, 22, 21, 53, 0, 0, time.UTC), nil
}

var _ = Describe("account", func() {
	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	addrs := trinary.Hashes{
		"SZMPOVAATOJCDNMPDEWFWSWVHINRFCQKM9RJRXHMLEPQJAICVWAQPOIEFODWUFYETQNGVGTLFQXRRHWUDVAEAZEJTY",
		"XUJGATAMJXVV9ZTAJUHUDSFTXPKNOAOZOBDXTDQQYXDIYYAXLXWNXQHEAYJOFHC9YXMVSFAGQQEDVREAXMKQOFFFKD",
		"RAEZOG9L9CWBCPZSYR9IZ9SUOERVKPXLJBAVYMPBGZKXIFIENMNXCNFBRZPG9QKPN9GOOQSZTIAMIISNWJZNGBFUOZ",
		"TYQPXKRGOWFLAIOZNBLZOBYLXCQCHJCWNPGBELXDHNQPODE9ZWURBVMUOVTPSVCTRNVTLKIEJDFBII9ACYT9RVBOWZ",
		"SKEPET9S9KNDHLHRTZQURARNYYACICNOVIVQNUQSSURUNKTEAGCSB9SAAZFJEDBAHLRSA9DTIMLF9IJMDMFPPIMPGA",
		"EGHTBJLOOUIMSHNYZEUZNJSJLNLZBIOWFOEMVZIGDLVCFPLVETGBKVOBUKLHBXMMOFPLNCVVPPCGAANQD99UVQSBLW",
		"AJAEMMP9RM9ZPWCTTAEKZPU9XCDZVNHTAHKNWWX9SORFPEXYFBHFPWVTRLEIQVZYWXBLGXJZDOGXKQLGZMWSTCUHNC",
		"XSCRTQVSEYMSZQEGHGAQMMYOJKZRLEVKRFOOGNCDAXDCVDPELAHJPSFSVULBTENPXIVZRYFDXUXYTPJXXRDYGAXBAX",
		"AKL9CZJLCFZWECGGDACSKDWWEVABNEAZAWKCQXGZYXIVQOZZDJYIPXTVS9AZBPFVCWYPCEUDYAQFJSRLAUZUD99ZPA",
		"LUENADZTQWWUQ9P9EXINSJJVNZLYVARHYMEY9QMNFZNF9IVMBBMKGUBSZPNMTEMMKVWTMUVLQUBHCHHHXFACPHILHA",
	}

	addrsWC := trinary.Hashes{
		"SZMPOVAATOJCDNMPDEWFWSWVHINRFCQKM9RJRXHMLEPQJAICVWAQPOIEFODWUFYETQNGVGTLFQXRRHWUD",
		"XUJGATAMJXVV9ZTAJUHUDSFTXPKNOAOZOBDXTDQQYXDIYYAXLXWNXQHEAYJOFHC9YXMVSFAGQQEDVREAX",
		"RAEZOG9L9CWBCPZSYR9IZ9SUOERVKPXLJBAVYMPBGZKXIFIENMNXCNFBRZPG9QKPN9GOOQSZTIAMIISNW",
		"TYQPXKRGOWFLAIOZNBLZOBYLXCQCHJCWNPGBELXDHNQPODE9ZWURBVMUOVTPSVCTRNVTLKIEJDFBII9AC",
		"SKEPET9S9KNDHLHRTZQURARNYYACICNOVIVQNUQSSURUNKTEAGCSB9SAAZFJEDBAHLRSA9DTIMLF9IJMD",
		"EGHTBJLOOUIMSHNYZEUZNJSJLNLZBIOWFOEMVZIGDLVCFPLVETGBKVOBUKLHBXMMOFPLNCVVPPCGAANQD",
		"AJAEMMP9RM9ZPWCTTAEKZPU9XCDZVNHTAHKNWWX9SORFPEXYFBHFPWVTRLEIQVZYWXBLGXJZDOGXKQLGZ",
		"XSCRTQVSEYMSZQEGHGAQMMYOJKZRLEVKRFOOGNCDAXDCVDPELAHJPSFSVULBTENPXIVZRYFDXUXYTPJXX",
		"AKL9CZJLCFZWECGGDACSKDWWEVABNEAZAWKCQXGZYXIVQOZZDJYIPXTVS9AZBPFVCWYPCEUDYAQFJSRLA",
		"LUENADZTQWWUQ9P9EXINSJJVNZLYVARHYMEY9QMNFZNF9IVMBBMKGUBSZPNMTEMMKVWTMUVLQUBHCHHHX",
	}

	var acc account.Account
	var em event.EventMachine
	var transferPoller *poller.TransferPoller
	var st *inmemory.InMemoryStore

	newAccount := func() {
		st = inmemory.NewInMemoryStore()
		em = event.NewEventMachine()
		b := builder.NewBuilder().
			WithAPI(api).
			WithStore(st).
			WithDepth(3).
			WithMWM(9).
			WithSecurityLevel(usedSecLvl).
			WithTimeSource(&fakeclock{}).
			WithEvents(em).
			WithSeed(seed)

		transferPoller = poller.NewTransferPoller(b.Settings(), poller.NewPerTailReceiveEventFilter(), 0)
		acc, err = b.Build(transferPoller)
		if err != nil {
			panic(err)
		}
		acc.Start()
	}

	Context("account creation", func() {
		It("successfully creates accounts given valid parameters/options", func() {
			newAccount()
			Expect(err).ToNot(HaveOccurred())
			Expect(acc.IsNew()).To(BeTrue())
			Expect(acc.Shutdown()).ToNot(HaveOccurred())
		})
	})

	Context("Addresses", func() {
		BeforeEach(newAccount)

		Context("NewDepositAddress()", func() {
			It("should return the correct address", func() {
				for _, compareAddr := range addrs {
					t := time.Now().AddDate(0, 0, 1)
					depositAddr, err := acc.AllocateDepositAddress(&deposit.Conditions{TimeoutAt: &t})
					Expect(err).ToNot(HaveOccurred())
					Expect(depositAddr.Address).To(Equal(compareAddr))
				}
			})
		})
	})

	Context("Operations", func() {
		Context("AvailableBalance()", func() {
			BeforeEach(newAccount)

			It("returns the correct usable balance", func() {
				defer gock.Flush()
				milestoneHash := strings.Repeat("M", 81)
				gock.New(DefaultLocalIRIURI).
					Post("/").
					MatchType("json").
					JSON(GetNodeInfoCommand{Command: Command{GetNodeInfoCmd}}).
					Reply(200).
					JSON(GetNodeInfoResponse{
						AppName:                            "IRI",
						AppVersion:                         "",
						Duration:                           100,
						JREAvailableProcessors:             4,
						JREFreeMemory:                      13020403,
						JREMaxMemory:                       1241331231,
						JRETotalMemory:                     4245234332,
						LatestMilestone:                    milestoneHash,
						LatestMilestoneIndex:               1,
						LatestSolidSubtangleMilestone:      milestoneHash,
						LatestSolidSubtangleMilestoneIndex: 1,
						Neighbors:                          5,
						PacketsQueueSize:                   23,
						Time:                               213213214,
						Tips:                               123,
						TransactionsToRequest:              10,
					})

				// the ordering of the addresses towards getBalance is random, as the stored deposit
				// addresses are in a map and thereby ordering is not preserved.
				permutations := []trinary.Hashes{{addrsWC[1], addrsWC[0]}, {addrsWC[0], addrsWC[1]}}
				for _, slice := range permutations {
					gock.New(DefaultLocalIRIURI).
						Persist().
						Post("/").
						MatchType("json").
						JSON(GetBalancesCommand{
							Command:   Command{GetBalancesCmd},
							Addresses: slice,
							Tips:      trinary.Hashes{milestoneHash},
						}).
						Reply(200).
						JSON(GetBalancesResponse{
							Duration:       100,
							Balances:       []string{"10", "20"},
							Milestone:      strings.Repeat("M", 81),
							MilestoneIndex: 1,
						})
				}
				t := time.Now().AddDate(0, 0, 1)
				for i := 0; i < 2; i++ {
					conds, err := acc.AllocateDepositAddress(&deposit.Conditions{TimeoutAt: &t})
					Expect(err).ToNot(HaveOccurred())
					Expect(conds.Address).To(Equal(addrs[i]))
				}
				balance, err := acc.AvailableBalance()
				Expect(err).ToNot(HaveOccurred())
				Expect(balance).To(Equal(uint64(30)))
			})

			It("returns 0 available balance on a new account", func() {
				balance, err := acc.AvailableBalance()
				Expect(err).ToNot(HaveOccurred())
				Expect(balance).To(Equal(uint64(0)))
			})
		})

		Context("Send()", func() {
			BeforeEach(newAccount)

			It("sends funds, fire events and completes with the correct state", func(done Done) {
				const targetAddr = "JURSJVFIECKJYEHPATCXADQGHABKOOEZCRUHLIDHPNPIGRCXBFBWVISWCF9ODWQKLXBKY9FACCKVXRAGZXKUQYNIY9"
				const targetAddrWC = "JURSJVFIECKJYEHPATCXADQGHABKOOEZCRUHLIDHPNPIGRCXBFBWVISWCF9ODWQKLXBKY9FACCKVXRAGZ"
				milestoneHash := strings.Repeat("M", 81)

				expectedPreparedTrytes := []trinary.Trytes{
					"F9FTLT9OCMACNHUMRSCCAWLSDGSHIZTRGYWGWIULQZWVGPA9YR9XQPSQBUBIAVBAXZVMDZIOHZOYECXNABPYCYHPUJXYMFSTRUYHT9LPNKV9KVZTLASSFWABZFUYXGLAWY9QGVLRGD9DRXAGJTMEB9SEHNPVHGYFYZR9FRMGXGKWICJCOGIVVXDHQZYXO9FYMIXRWHSGFCVB9Y9JC9JHUQZQLGOJTRIWAPPVHCGTNHNXHYTKAVWABYPHWTPBFUJDNSXZZUAJNVLUNYCDDTVIDNVYWZOC9UYIGBXTZEKSDGVODGPDUPDRGK9ITQINJUVEHREDUDYDM9GPMPUUGDYYACYYYUNQLZTUBSUTDY9XNJEGUCVQLEYVQU9NROORJVAIKDVODH9NGB9OOYELEFEXBGZ9ONNRDPVADUVNTTVIATO9SGJLLUUABFTPURFVZGUZRDFSGJCBZCDRIJYLILTWRSIXDGFKDCPSSWMJZXUOMNPDKMSYBOUXGBQ9DURPCXULZIMRPGVYSPNZDSNHFFJNUWLEW9IEOYTOSUEUSBNPIZZQQJMVQVQL9ICVTSZCXLLBEDBKZEMDPSKLTBNIXMBIBJEBOOZBKXTIOXCOKTWDZUAZZVYRGKHPNHIAYQNSAXKI9ODSUCAXXDMKRKWMFWAFRVMSYD9IGPNJLAADSSGUWF9IKFAJRGLWXYTIRMZBQLOFOYKGTYVGBOFHRUPRQRPCNSMYXGWVAAB9RMRYGXCQVQOCJSDNIEJIHBCNXKJJNTTJWCCMWNLJSAYQBSGAGBEHRYLKYQOAYHVZKZWGGQ9XKZWXUPLEBSIVHFODJB9YRXLNATBFIDPMEMOKDFEOUYYLJLNFHCOFQ9VAUGTBDZQWQXJFMQSHMAONFWFIKPWMINPEBT9GNGTJRQWJTZYEVLYIKDUTBGVKVKQSDODENM9KXDJSWCSDX9CHNCLBRDNZWJS99EFST9JQLTWDRQR9YKUYRVWMOLFHFKPTSSHOLTBZSPSMMHZIGTPL9BFXURBCDIK9SHIWJUKFEYMSJDKLBVWNEJ9WRKEDADXTXFBQOY9XDDWNUUXETVR9BHXVIGZEIEPALJHGMRPKQYCTJCGTIKBNPTGN9LHSNOUDPKOD9HVYMFKZX9AKNTCALRSYOVQBV9B9MX9CIEEGVJB9YWBIUMAPIETEFEGAQIP9AJVUIPQJAHBBJBZBZIBHSUSGRTKHRMWH9TQKQGULKFMQKJKGMDZVYRJ9LNOEFKNTNTEYUK9YXSYNMDJGSDYOOGBJBVOQTCMTTJOXY99UTGQKXPGY9GESYQJVJYC999SYYTYCZN9DCPQHTEUSLU9XRZTVEDMDVMCWQLCDNBP9YO9TMIXAYFIJFDUOCHDOXFLXYNEO9OKYKBGXDNPTJDH9CDNGDNYKXHTUENOADBOA9HJKRVYWF9WYWZ9FLQWGLUENETWTBSPTKCVBXYGIOZQHEHMSTEQKKOASHOYIWQEVMPLHNMDDIOGTWHKV9FUXWZYWCRWOYOCXPVMLVFHIPYAFAIXBBMZVDPQOAJBITVIUYBWOIMLIOWHLQE9CDCDVMEYIBLPKDVPUBZJQHTSIIYDSDXESLXTTWYPBKANFGP9AKYYJLHBK9XZANKJKFGBEOUDYRIXFRCJNUWMIUQWTPLGQPVXRRLAVVIFZTFDVOKXGJUWYBGETTNGRURBCYXRLNBMVFRWPDLULXSYPIMIUAEZ9DHINJERPZ9WVKVRBQWAJIOIONIHWHHMWZNEICIUHWZOQ9JTWDRTECJLUDTXGPPYQGCI99VTGHIDXTLUILOIRKHCNHZZRVMPOHVUWJEMBPONPALCHOOCMBHDFUHJOJVNEKNWJONXTNUYHFXSBXD9MCBNAMDCRACGTQWI9OBVXEKRFCZUWTMRSMQQVAOUOHSMQOFCPCCJFGAELSULMHXNAHXWOKBEKRDCTPNKDQQXQDRHQSWNTQ9IDPGJISA9PZSULUGYVNGOFKFXRRCQLIZ9LKCHOQ99KRISZOXFDZTTTWNXPKIDIJAAGASSTMJMNQ9HGZZYRFYBACAXAEMBCNWWIWCJJQZTDQUXQEVCSBEFYFAKVPLLNRMPJKHKNOLATZABFJLMBPWAQCEJCVWXJCGVTUTFYLFLYARPSZSQQJGFAPZYYAKKQMULPLMHXWQCNBTXWHAZCWFSY9KJSEEDIUMGPFWASZMPOVAATOJCDNMPDEWFWSWVHINRFCQKM9RJRXHMLEPQJAICVWAQPOIEFODWUFYETQNGVGTLFQXRRHWUDHW9999999999999999999999999999999999999999999999999999UGQHN9D99A99999999A999999999PKJIBTGLBIIVJ9AREWCTIUUURUHINTWNCFYFGEJ9XIPGVCRBVCGGXJOJFBSJVZRZDLAQSOGQEWGDKCWW999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999",
					"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999JURSJVFIECKJYEHPATCXADQGHABKOOEZCRUHLIDHPNPIGRCXBFBWVISWCF9ODWQKLXBKY9FACCKVXRAGZSD9999999999999999999999999OL9999999999999999999999999UGQHN9D99999999999A999999999PKJIBTGLBIIVJ9AREWCTIUUURUHINTWNCFYFGEJ9XIPGVCRBVCGGXJOJFBSJVZRZDLAQSOGQEWGDKCWW999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999",
				}

				powedPreparedTrytes := []trinary.Trytes{
					"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999JURSJVFIECKJYEHPATCXADQGHABKOOEZCRUHLIDHPNPIGRCXBFBWVISWCF9ODWQKLXBKY9FACCKVXRAGZSD9999999999999999999999999OL9999999999999999999999999UGQHN9D99999999999A999999999PKJIBTGLBIIVJ9AREWCTIUUURUHINTWNCFYFGEJ9XIPGVCRBVCGGXJOJFBSJVZRZDLAQSOGQEWGDKCWWOWXEPBIFBS9FJEPWFOZOZJAKMURCGDSEOVFGDRGPCMXUZUXMYUIQAXEGQBDMPIKYRLRLBUBSEYRMUI999AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA999999999999999999999999999OWFXNRNME999999999K99999999HBB9999999UJ999999999999999",
					"F9FTLT9OCMACNHUMRSCCAWLSDGSHIZTRGYWGWIULQZWVGPA9YR9XQPSQBUBIAVBAXZVMDZIOHZOYECXNABPYCYHPUJXYMFSTRUYHT9LPNKV9KVZTLASSFWABZFUYXGLAWY9QGVLRGD9DRXAGJTMEB9SEHNPVHGYFYZR9FRMGXGKWICJCOGIVVXDHQZYXO9FYMIXRWHSGFCVB9Y9JC9JHUQZQLGOJTRIWAPPVHCGTNHNXHYTKAVWABYPHWTPBFUJDNSXZZUAJNVLUNYCDDTVIDNVYWZOC9UYIGBXTZEKSDGVODGPDUPDRGK9ITQINJUVEHREDUDYDM9GPMPUUGDYYACYYYUNQLZTUBSUTDY9XNJEGUCVQLEYVQU9NROORJVAIKDVODH9NGB9OOYELEFEXBGZ9ONNRDPVADUVNTTVIATO9SGJLLUUABFTPURFVZGUZRDFSGJCBZCDRIJYLILTWRSIXDGFKDCPSSWMJZXUOMNPDKMSYBOUXGBQ9DURPCXULZIMRPGVYSPNZDSNHFFJNUWLEW9IEOYTOSUEUSBNPIZZQQJMVQVQL9ICVTSZCXLLBEDBKZEMDPSKLTBNIXMBIBJEBOOZBKXTIOXCOKTWDZUAZZVYRGKHPNHIAYQNSAXKI9ODSUCAXXDMKRKWMFWAFRVMSYD9IGPNJLAADSSGUWF9IKFAJRGLWXYTIRMZBQLOFOYKGTYVGBOFHRUPRQRPCNSMYXGWVAAB9RMRYGXCQVQOCJSDNIEJIHBCNXKJJNTTJWCCMWNLJSAYQBSGAGBEHRYLKYQOAYHVZKZWGGQ9XKZWXUPLEBSIVHFODJB9YRXLNATBFIDPMEMOKDFEOUYYLJLNFHCOFQ9VAUGTBDZQWQXJFMQSHMAONFWFIKPWMINPEBT9GNGTJRQWJTZYEVLYIKDUTBGVKVKQSDODENM9KXDJSWCSDX9CHNCLBRDNZWJS99EFST9JQLTWDRQR9YKUYRVWMOLFHFKPTSSHOLTBZSPSMMHZIGTPL9BFXURBCDIK9SHIWJUKFEYMSJDKLBVWNEJ9WRKEDADXTXFBQOY9XDDWNUUXETVR9BHXVIGZEIEPALJHGMRPKQYCTJCGTIKBNPTGN9LHSNOUDPKOD9HVYMFKZX9AKNTCALRSYOVQBV9B9MX9CIEEGVJB9YWBIUMAPIETEFEGAQIP9AJVUIPQJAHBBJBZBZIBHSUSGRTKHRMWH9TQKQGULKFMQKJKGMDZVYRJ9LNOEFKNTNTEYUK9YXSYNMDJGSDYOOGBJBVOQTCMTTJOXY99UTGQKXPGY9GESYQJVJYC999SYYTYCZN9DCPQHTEUSLU9XRZTVEDMDVMCWQLCDNBP9YO9TMIXAYFIJFDUOCHDOXFLXYNEO9OKYKBGXDNPTJDH9CDNGDNYKXHTUENOADBOA9HJKRVYWF9WYWZ9FLQWGLUENETWTBSPTKCVBXYGIOZQHEHMSTEQKKOASHOYIWQEVMPLHNMDDIOGTWHKV9FUXWZYWCRWOYOCXPVMLVFHIPYAFAIXBBMZVDPQOAJBITVIUYBWOIMLIOWHLQE9CDCDVMEYIBLPKDVPUBZJQHTSIIYDSDXESLXTTWYPBKANFGP9AKYYJLHBK9XZANKJKFGBEOUDYRIXFRCJNUWMIUQWTPLGQPVXRRLAVVIFZTFDVOKXGJUWYBGETTNGRURBCYXRLNBMVFRWPDLULXSYPIMIUAEZ9DHINJERPZ9WVKVRBQWAJIOIONIHWHHMWZNEICIUHWZOQ9JTWDRTECJLUDTXGPPYQGCI99VTGHIDXTLUILOIRKHCNHZZRVMPOHVUWJEMBPONPALCHOOCMBHDFUHJOJVNEKNWJONXTNUYHFXSBXD9MCBNAMDCRACGTQWI9OBVXEKRFCZUWTMRSMQQVAOUOHSMQOFCPCCJFGAELSULMHXNAHXWOKBEKRDCTPNKDQQXQDRHQSWNTQ9IDPGJISA9PZSULUGYVNGOFKFXRRCQLIZ9LKCHOQ99KRISZOXFDZTTTWNXPKIDIJAAGASSTMJMNQ9HGZZYRFYBACAXAEMBCNWWIWCJJQZTDQUXQEVCSBEFYFAKVPLLNRMPJKHKNOLATZABFJLMBPWAQCEJCVWXJCGVTUTFYLFLYARPSZSQQJGFAPZYYAKKQMULPLMHXWQCNBTXWHAZCWFSY9KJSEEDIUMGPFWASZMPOVAATOJCDNMPDEWFWSWVHINRFCQKM9RJRXHMLEPQJAICVWAQPOIEFODWUFYETQNGVGTLFQXRRHWUDHW9999999999999999999999999999999999999999999999999999UGQHN9D99A99999999A999999999PKJIBTGLBIIVJ9AREWCTIUUURUHINTWNCFYFGEJ9XIPGVCRBVCGGXJOJFBSJVZRZDLAQSOGQEWGDKCWWAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB999999999999999999999999999IVFXNRNME999999999K99999999ARB99999999A999999999999999",
				}

				finalTxs, err := transaction.AsTransactionObjects(powedPreparedTrytes, nil)
				Expect(err).ToNot(HaveOccurred())

				trunkTx := strings.Repeat("A", 81)
				branchTx := strings.Repeat("B", 81)

				// availableBalance query
				defer gock.Flush()
				gock.New(DefaultLocalIRIURI).
					Persist().
					Post("/").
					MatchType("json").
					JSON(GetBalancesCommand{
						Command:   Command{GetBalancesCmd},
						Addresses: addrsWC[:1],
						Tips:      trinary.Hashes{milestoneHash},
					}).
					Reply(200).
					JSON(GetBalancesResponse{
						Duration:       100,
						Balances:       []string{"100"},
						Milestone:      strings.Repeat("M", 81),
						MilestoneIndex: 1,
					})

				// has transfer query
				gock.New(DefaultLocalIRIURI).
					Persist().
					Post("/").
					MatchType("json").
					JSON(FindTransactionsCommand{Command: Command{FindTransactionsCmd}, FindTransactionsQuery: FindTransactionsQuery{
						Addresses: addrsWC[:1],
					}}).
					Reply(200).
					JSON(FindTransactionsResponse{Hashes: trinary.Hashes{}})

				gock.New(DefaultLocalIRIURI).
					Post("/").
					MatchType("json").
					JSON(GetTransactionsToApproveCommand{
						Command: Command{GetTransactionsToApproveCmd}, Depth: 3,
					}).
					Reply(200).
					JSON(GetTransactionsToApproveResponse{
						Duration: 100,
						TransactionsToApprove: TransactionsToApprove{
							TrunkTransaction:  trunkTx,
							BranchTransaction: branchTx,
						},
					})

				gock.New(DefaultLocalIRIURI).
					Post("/").
					MatchType("json").
					JSON(AttachToTangleCommand{
						Command: Command{AttachToTangleCmd}, TrunkTransaction: trunkTx, BranchTransaction: branchTx,
						MinWeightMagnitude: 9, Trytes: expectedPreparedTrytes,
					}).
					Reply(200).
					JSON(AttachToTangleResponse{Trytes: powedPreparedTrytes})

				gock.New(DefaultLocalIRIURI).
					Post("/").
					MatchType("json").
					JSON(StoreTransactionsCommand{Command: Command{StoreTransactionsCmd}, Trytes: powedPreparedTrytes}).
					Reply(200)

				gock.New(DefaultLocalIRIURI).
					Post("/").
					MatchType("json").
					JSON(BroadcastTransactionsCommand{Command: Command{BroadcastTransactionsCmd}, Trytes: powedPreparedTrytes}).
					Reply(200)

				gock.New(DefaultLocalIRIURI).
					Persist().
					Post("/").
					MatchType("json").
					JSON(GetNodeInfoCommand{Command: Command{GetNodeInfoCmd}}).
					Reply(200).
					JSON(GetNodeInfoResponse{
						AppName:                            "IRI",
						AppVersion:                         "",
						Duration:                           100,
						JREAvailableProcessors:             4,
						JREFreeMemory:                      13020403,
						JREMaxMemory:                       1241331231,
						JRETotalMemory:                     4245234332,
						LatestMilestone:                    milestoneHash,
						LatestMilestoneIndex:               1,
						LatestSolidSubtangleMilestone:      milestoneHash,
						LatestSolidSubtangleMilestoneIndex: 1,
						Neighbors:                          5,
						PacketsQueueSize:                   23,
						Time:                               213213214,
						Tips:                               123,
						TransactionsToRequest:              10,
					})

				gock.New(DefaultLocalIRIURI).
					Persist().
					Post("/").
					MatchType("json").
					JSON(WereAddressesSpentFromCommand{
						Command:   Command{WereAddressesSpentFromCmd},
						Addresses: trinary.Hashes{targetAddrWC},
					}).
					Reply(200).
					JSON(WereAddressesSpentFromResponse{
						States: []bool{false},
					})

				// hypothetical deposit address given to someone and got some funds
				t := time.Now().AddDate(0, 0, 1)
				expectedAmount := uint64(100)
				_, err = acc.AllocateDepositAddress(&deposit.Conditions{TimeoutAt: &t, ExpectedAmount: &expectedAmount})
				Expect(err).ToNot(HaveOccurred())

				By("having the correct current usable balance")
				balance, err := acc.AvailableBalance()
				Expect(err).ToNot(HaveOccurred())
				Expect(balance).To(Equal(uint64(100)))

				By("registering an event listener for sending transfers")
				eventListener := listener.NewChannelEventListener(em).RegSentTransfers()
				resultBackCheck := make(chan bundle.Bundle)

				transferPoller.Poll()

				go func() {
					bndl := <-eventListener.SentTransfer
					resultBackCheck <- bndl
				}()

				By("giving a valid target address and transfer amount")
				bndl, err := acc.Send(account.Recipient{Address: targetAddr, Value: 100})
				Expect(err).ToNot(HaveOccurred())
				Expect(bndl).To(Equal(finalTxs))

				bundleFromEvent := <-resultBackCheck
				Expect(bundleFromEvent).To(Equal(finalTxs))

				gock.New(DefaultLocalIRIURI).
					Post("/").
					MatchType("json").
					JSON(GetInclusionStatesCommand{
						Command:      Command{GetInclusionStatesCmd},
						Transactions: trinary.Hashes{finalTxs[0].Hash},
					}).
					Reply(200).
					JSON(GetInclusionStatesResponse{States: []bool{true}})

				for i := range finalTxs {
					gock.New(DefaultLocalIRIURI).
						Post("/").
						MatchType("json").
						JSON(GetTrytesCommand{Command: Command{GetTrytesCmd}, Hashes: trinary.Hashes{finalTxs[i].Hash}}).
						Reply(200).
						JSON(AttachToTangleResponse{Trytes: []trinary.Trytes{powedPreparedTrytes[i]}})
				}

				By("registering an event listener for confirmed sent transfers")
				eventListener.RegConfirmedTransfers()

				go func() {
					transferPoller.Poll()
				}()

				bundleFromEvent = <-eventListener.TransferConfirmed
				Expect(bundleFromEvent[0].Bundle).To(Equal(finalTxs[0].Bundle))

				state, err := st.LoadAccount(id)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(state.PendingTransfers)).To(Equal(0))

				close(done)
			}, 5)
		})

	})

})
