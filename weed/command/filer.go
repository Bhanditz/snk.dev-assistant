package command

import (
	"net/http"
	"strconv"
	"time"

	"github.com/chrislusf/seaweedfs/weed/filer2"
	"github.com/chrislusf/seaweedfs/weed/glog"
	"github.com/chrislusf/seaweedfs/weed/pb/filer_pb"
	"github.com/chrislusf/seaweedfs/weed/server"
	"github.com/chrislusf/seaweedfs/weed/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"strings"
)

var (
	f FilerOptions
)

type FilerOptions struct {
	masters                 *string
	ip                      *string
	port                    *int
	grpcPort                *int
	publicPort              *int
	collection              *string
	defaultReplicaPlacement *string
	redirectOnRead          *bool
	disableDirListing       *bool
	maxMB                   *int
	secretKey               *string
}

func init() {
	cmdFiler.Run = runFiler // break init cycle
	f.masters = cmdFiler.Flag.String("master", "localhost:9333", "comma-separated master servers")
	f.collection = cmdFiler.Flag.String("collection", "", "all data will be stored in this collection")
	f.ip = cmdFiler.Flag.String("ip", "", "filer server http listen ip address")
	f.port = cmdFiler.Flag.Int("port", 8888, "filer server http listen port")
	f.grpcPort = cmdFiler.Flag.Int("port.grpc", 0, "filer grpc server listen port, default to http port + 10000")
	f.publicPort = cmdFiler.Flag.Int("port.public", 0, "port opened to public")
	f.defaultReplicaPlacement = cmdFiler.Flag.String("defaultReplicaPlacement", "000", "default replication type if not specified")
	f.redirectOnRead = cmdFiler.Flag.Bool("redirectOnRead", false, "whether proxy or redirect to volume server during file GET request")
	f.disableDirListing = cmdFiler.Flag.Bool("disableDirListing", false, "turn off directory listing")
	f.maxMB = cmdFiler.Flag.Int("maxMB", 32, "split files larger than the limit")
	f.secretKey = cmdFiler.Flag.String("secure.secret", "", "secret to encrypt Json Web Token(JWT)")
}

var cmdFiler = &Command{
	UsageLine: "filer -port=8888 -master=<ip:port>[,<ip:port>]*",
	Short:     "start a file server that points to a master server, or a list of master servers",
	Long: `start a file server which accepts REST operation for any files.

	//create or overwrite the file, the directories /path/to will be automatically created
	POST /path/to/file
	//get the file content
	GET /path/to/file
	//create or overwrite the file, the filename in the multipart request will be used
	POST /path/to/
	//return a json format subdirectory and files listing
	GET /path/to/

	The configuration file "filer.toml" is read from ".", "$HOME/.seaweedfs/", or "/etc/seaweedfs/", in that order.

	The following are example filer.toml configuration file.

` + filer2.FILER_TOML_EXAMPLE + "\n",
}

func runFiler(cmd *Command, args []string) bool {

	f.start()

	return true
}

func (fo *FilerOptions) start() {

	defaultMux := http.NewServeMux()
	publicVolumeMux := defaultMux

	if *fo.publicPort != 0 {
		publicVolumeMux = http.NewServeMux()
	}

	masters := *f.masters

	fs, nfs_err := weed_server.NewFilerServer(defaultMux, publicVolumeMux,
		*fo.ip, *fo.port, strings.Split(masters, ","), *fo.collection,
		*fo.defaultReplicaPlacement, *fo.redirectOnRead, *fo.disableDirListing,
		*fo.maxMB,
		*fo.secretKey,
	)
	if nfs_err != nil {
		glog.Fatalf("Filer startup error: %v", nfs_err)
	}

	if *fo.publicPort != 0 {
		publicListeningAddress := *fo.ip + ":" + strconv.Itoa(*fo.publicPort)
		glog.V(0).Infoln("Start Seaweed filer server", util.VERSION, "public at", publicListeningAddress)
		publicListener, e := util.NewListener(publicListeningAddress, 0)
		if e != nil {
			glog.Fatalf("Filer server public listener error on port %d:%v", *fo.publicPort, e)
		}
		go func() {
			if e := http.Serve(publicListener, publicVolumeMux); e != nil {
				glog.Fatalf("Volume server fail to serve public: %v", e)
			}
		}()
	}

	glog.V(0).Infoln("Start Seaweed Filer", util.VERSION, "at port", strconv.Itoa(*fo.port))
	filerListener, e := util.NewListener(
		":"+strconv.Itoa(*fo.port),
		time.Duration(10)*time.Second,
	)
	if e != nil {
		glog.Fatalf("Filer listener error: %v", e)
	}

	// starting grpc server
	grpcPort := *f.grpcPort
	if grpcPort == 0 {
		grpcPort = *f.port + 10000
	}
	grpcL, err := util.NewListener(":"+strconv.Itoa(grpcPort), 0)
	if err != nil {
		glog.Fatalf("failed to listen on grpc port %d: %v", grpcPort, err)
	}
	grpcS := grpc.NewServer()
	filer_pb.RegisterSeaweedFilerServer(grpcS, fs)
	reflection.Register(grpcS)
	go grpcS.Serve(grpcL)

	httpS := &http.Server{Handler: defaultMux}
	if err := httpS.Serve(filerListener); err != nil {
		glog.Fatalf("Filer Fail to serve: %v", e)
	}

}
