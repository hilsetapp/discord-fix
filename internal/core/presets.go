package core

// Стратегии обхода. Разные провайдеры режут по-разному, поэтому их несколько,
// а режим «Авто» перебирает все по очереди, пока дискорд не поднимется.
//
// Работает только zapret (winws): списки сайтов и UDP-голос он умеет из коробки.
// GoodbyeDPI - запасной простой режим, UDP он не трогает (голосовые каналы
// в нём могут не завестись, приложение честно об этом предупреждает).

type Engine string

const (
	EngineZapret Engine = "zapret"
	EngineGdpi   Engine = "gdpi"
)

type Preset struct {
	ID      string
	Name    string // как показываем человеку
	Short   string // одна строка подсказки
	Engine  Engine
	tcpArgs []string // что вставить в цепочку --filter-tcp=443
}

// общие куски winws для всех стратегий
var udpQuic = []string{
	"--filter-udp=443",
	"--dpi-desync=fake", "--dpi-desync-repeats=6",
	"--dpi-desync-fake-quic=quic_initial_www_google_com.bin",
}

var udpVoice = []string{
	"--filter-udp=50000-50099",
	"--dpi-desync=fake", "--dpi-desync-repeats=6",
}

// Таблица стратегий - собрано по открытым конфигам из zapret (bol-van)
// и популярных полумашинных конфигураций сообщества.
var presets = []Preset{
	{
		ID: "std", Name: "Стандарт", Short: "fake + split2, подходит большинству",
		Engine: EngineZapret,
		tcpArgs: []string{
			"--dpi-desync=fake,split2",
			"--dpi-desync-autottl=2",
			"--dpi-desync-fooling=md5sig",
			"--dpi-desync-fake-tls=tls_clienthello_www_google_com.bin",
		},
	},
	{
		ID: "autosplit", Name: "Автосплит", Short: "режет клиентское приветствие на куски",
		Engine: EngineZapret,
		tcpArgs: []string{
			"--dpi-desync=autosplit",
			"--dpi-desync-split-pos=1",
			"--dpi-desync-fooling=badseq",
			"--dpi-desync-repeats=6",
		},
	},
	{
		ID: "multi", Name: "Мульти-фейк", Short: "фейковый hello от google + мультисплит",
		Engine: EngineZapret,
		tcpArgs: []string{
			"--dpi-desync=fake,multisplit",
			"--dpi-desync-split-pos=1",
			"--dpi-desync-fooling=md5sig",
			"--dpi-desync-fake-tls=tls_clienthello_www_google_com.bin",
		},
	},
	{
		ID: "mgts", Name: "Провайдер-стиль", Short: "для придирчивых сетей (МГТС и похожие)",
		Engine: EngineZapret,
		tcpArgs: []string{
			"--dpi-desync=fake,fakedsplit",
			"--dpi-desync-ttl=4",
			"--dpi-desync-split-pos=1",
			"--dpi-desync-fooling=badsum,md5sig",
			"--dpi-desync-fake-tls=tls_clienthello_sberbank_ru.bin",
		},
	},
	{
		ID: "gdpi", Name: "Запасной (GoodbyeDPI)", Short: "простой режим, без голосовых каналов",
		Engine:  EngineGdpi,
		tcpArgs: []string{"-5"},
	},
}

// Presets отдаёт копию таблицы стратегий.
func Presets() []Preset {
	out := make([]Preset, len(presets))
	copy(out, presets)
	return out
}

// FindPreset ищет стратегию по id, auto отдаёт nil.
func FindPreset(id string) *Preset {
	for i := range presets {
		if presets[i].ID == id {
			return &presets[i]
		}
	}
	return nil
}

// ArgsFor собирает полную командную строку для движка.
// hostlist - абсолютный путь к списку доменов.
// Чем цепочки у winws разделены флагом --new: tcp, потом udp quic, потом голос.
func ArgsFor(p *Preset, hostlist string) (exe string, args []string) {
	if p.Engine == EngineGdpi {
		return "goodbyedpi.exe", append([]string{}, p.tcpArgs...)
	}
	args = []string{"--wf-tcp=443", "--wf-udp=443,50000-50099"}
	args = append(args, "--filter-tcp=443", "--hostlist="+hostlist)
	args = append(args, p.tcpArgs...)
	args = append(args, "--new")
	args = append(args, udpQuic...)
	args = append(args, "--new")
	args = append(args, udpVoice...)
	return "winws.exe", args
}
