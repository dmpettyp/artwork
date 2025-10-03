package main

import (
	"context"
	"log/slog"
	"os"

	// "image"
	"log"

	"github.com/dmpettyp/artwork/application"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/dmpettyp/dorky"
	// "github.com/dmpettyp/pixelator/domain/node"
	// "github.com/dmpettyp/pixelator/lib/fileimagestore"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info("this is artwork")

	messageBus := dorky.NewMessageBus(logger)

	_, err := application.NewImageGraphCommandHandlers(messageBus)

	if err != nil {
		logger.Error("could not create image graph command handlers", "error", err)
		return
	}

	go messageBus.Start(context.Background())

	messageBus.Stop()

	// imageStore, err := fileimagestore.New("data/images")
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// sourceImage, err := imageStore.FetchImageByName("dave")
	//
	ig, err := imagegraph.NewImageGraph(
		imagegraph.MustNewImageGraphID(),
		"sweet new imagegraph",
	)

	if err != nil {
		log.Fatal(err)
	}

	nodeID1, err := imagegraph.ParseNodeID("10000000-f612-43a7-8ae3-07197a8e9e6f")

	if err != nil {
		log.Fatal(err)
	}

	err = ig.AddNode(nodeID1, imagegraph.NodeTypeInput, "My Input Node", "{}")

	if err != nil {
		log.Fatal(err)
	}

	err = ig.RemoveNode(nodeID1)

	if err != nil {
		log.Fatal(err)
	}

	err = ig.RemoveNode(nodeID1)

	if err != nil {
		log.Fatal(err)
	}

	// node1.Run(sourceImage)
	//
	// err = p.AddNode(
	// 	node1.GetID(),
	// 	node1.GetInputNames(),
	// 	node1.GetOutputNames(),
	// )
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// // nodeID2, err := node.ParseID("20000000-3e30-47ac-8eda-5b3f31f25d8e")
	//
	// // if err != nil {
	// // 		log.Fatal(err)
	// // }
	//
	// nodeID3, err := node.ParseID("30000000-3e30-47ac-8eda-5b3f31f25d8e")
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// node3, err := node.NewShrink(nodeID3, "")
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// err = node3.SetConfig(
	// 	node.Config{"max_width_height": 32},
	// )
	//
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	// fmt.Println("shrink to", node3.MaxWH)
	//
	// err = p.AddNode(
	// 	node3.GetID(),
	// 	node3.GetInputNames(),
	// 	node3.GetOutputNames(),
	// )
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// err = p.ConnectNodes(
	// 	nodeID1, "original",
	// 	nodeID3, "source",
	// )
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// nodeID4, err := node.ParseID("40000000-3e30-47ac-8eda-5b3f31f25d8e")
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// node4, err := node.NewResize(nodeID4, "")
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// err = node4.SetConfig(
	// 	node.Config{
	// 		"width": "dave",
	// 	},
	// )
	//
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	// err = node4.SetConfig(
	// 	node.Config{
	// 		"width": -1,
	// 	},
	// )
	//
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	// err = node4.SetConfig(
	// 	node.Config{
	// 		"width": uint(1000),
	// 	},
	// )
	//
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	// err = node4.SetConfig(
	// 	node.Config{
	// 		"width": 1000,
	// 	},
	// )
	//
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	// fmt.Println("width is", node4.Width)
	//
	// err = p.AddNode(
	// 	node4.GetID(),
	// 	node4.GetInputNames(),
	// 	node4.GetOutputNames(),
	// )
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// err = p.ConnectNodes(
	// 	nodeID3, "shrunk",
	// 	nodeID4, "source",
	// )
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// nodeID5, err := node.ParseID("50000000-3e30-47ac-8eda-5b3f31f25d8e")
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// node5, err := node.NewResizeTo(nodeID5, "resize to original")
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// err = node5.Run([]image.Image{sourceImage, sourceImage})
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// err = p.AddNode(
	// 	node5.GetID(),
	// 	node5.GetInputNames(),
	// 	node5.GetOutputNames(),
	// )
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// err = p.ConnectNodes(
	// 	nodeID3, "shrunk",
	// 	nodeID5, "source",
	// )
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// err = p.ConnectNodes(
	// 	nodeID1, "original",
	// 	nodeID5, "targetsize",
	// )
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// err = p.DisconnectNodes(
	// 	nodeID5, "targetsize",
	// )
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// // err = p.RemoveNode(nodeID2)
	// //
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	//
	// err = p.TriggerNode(nodeID3)
	//
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// // for _, e := range p.GetEvents() {
	// // 	fmt.Println(ddd.MessageToJsonString(e))
	// // }
	//
	// // err = p.RunNode(nodeID5)
	// //
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	//
	// // repo, err := jsonAdapter.NewPipelineRepository("data/json", imageStore)
	// //
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	// //
	// // err = repo.Add(p)
	// //
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	// //
	// // err = repo.Save()
	// //
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
}
