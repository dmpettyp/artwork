# artwork

artwork is used to model a pipeline for generating artwork. It creates an graph
of nodes, each of which has inputs and outputs that can be connected together
to process images in a non-destructive way.

## TODO

- implement different scaling types
- why is resize width/height float? make int
- publish node field config to frontend?
- resizeMatch - multiple resize options
- durable repository/unit of work implementation: sqlite

## Overview

The core model in this project is called an "ImageGraph". An ImageGraph
consists of a name, a version, and a collection of nodes which may be connected
to each other.

The ImageGraph is the primary way to interact with the pipeline, and provides
an interface that enables clients to manipulate the pipeline. This includes:
- adding nodes
- removing nodes
- connecting nodes
- disconnecting nodes
- setting node output images
- unsetting node output images

The ImageGraph model attempts to use domain driven design principles. As such, 
it does not have any dependencies that are specific to the implementation of
the service that is built around it. This includes storage (i.e. database) 
concerns, API concerns, or any other integrations that need to access or 
manipulated the ImageGraph.

The ImageGraph is primarily responsible for maintaining the Nodes that make
up the image pipeline. Nodes are modeled with:
- a name
- a version
- a type that indicates what transformation the node represents
- a generic configuration object that is used to configure the node's 
  image transoformation
- inputs which feed source images into the node for processing
- outputs which provide output images to be used as inputs for downstream nodes

The ImageGraph and Nodes are primarily used as a control plane to model and
configure the image pipeline. The actual image creation/manipulation is 
performed by processes external to the domain models, which set their output
in the domain models to drive further changes in the ImageGraph pipeline.



## Done

- NewImageGraph ✅
- AddNode ✅
- RemoveNode ✅
  - should unset image downstream if it is set ✅
- ConnectNodes ✅
- SetNodeOutputImage ✅
- UnsetNodeOutputImage ✅
- DisconnectNodes ✅
- Node configuration ✅
  - json blob, parse it and verify contents ✅
- Node Preview Image ✅
- Add a new node type that supports inputs!
- Inmem repository and unit of work ✅
- command, handlers and messagebus ✅
- regenerate events✅
- node states✅
- swagger documentation? ✅
- mapper ✅
- move config to actual json in the API layer ✅
- http APIlayer ✅
- change patch node/config to patch node to allow name changing ✅
  - add domain method to Node and ImageGraph ✅
  - add command ✅
  - modify patch endpoint to update multiple things ✅
  - ui - in progress ✅
- change UpdateUIMetadataCommand to not use a map, use a slice of structs ✅
- uploading images ✅
  - create depenedency (ImageStore) that implements interface to set and get images ✅
  - create handler that allows images to be uploaded and uses ImageStore the is injected ✅
- UI
  - show output as thumbnail ✅
  - drawer for inputs and outputs - not doing this, went with a table approach for inputs and outputs
  - output/input node green when set, red when not
- websocket implementation for events ✅
- message when outputs are being generated ✅
- output nodes should have side bar or something that shows the output images ✅
- more node types
  - get rid of scale
  - output ✅
  - blur ✅
  - resize ✅
  - resizeTo ✅
- split up viewport/layout ✅
- implement node preview...small images look crappy when scaled by the UI. but maybe SVG can help with that? ✅
- separate ws notifications for graph and layout changes ✅
- download functionality for output panel ✅
- extract current graph for bootstrap ✅
- ensure that input and output order is stable as defined in node type ✅
