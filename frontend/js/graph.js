// Graph state management with observable pattern

export class GraphState {
    constructor() {
        this.currentGraphId = null;
        this.currentGraph = null;
        this.listeners = new Set();
    }

    subscribe(listener) {
        this.listeners.add(listener);
        return () => this.listeners.delete(listener);
    }

    notify() {
        this.listeners.forEach(listener => listener(this.currentGraph));
    }

    setCurrentGraph(graph) {
        this.currentGraphId = graph?.id || null;
        this.currentGraph = graph;
        this.notify();
    }

    getCurrentGraph() {
        return this.currentGraph;
    }

    getCurrentGraphId() {
        return this.currentGraphId;
    }

    updateNode(nodeId, updates) {
        if (!this.currentGraph) return;

        const nodeIndex = this.currentGraph.nodes.findIndex(n => n.id === nodeId);
        if (nodeIndex !== -1) {
            this.currentGraph.nodes[nodeIndex] = {
                ...this.currentGraph.nodes[nodeIndex],
                ...updates
            };
            this.notify();
        }
    }

    addNode(node) {
        if (!this.currentGraph) return;

        this.currentGraph.nodes.push(node);
        this.notify();
    }

    removeNode(nodeId) {
        if (!this.currentGraph) return;

        this.currentGraph.nodes = this.currentGraph.nodes.filter(n => n.id !== nodeId);
        this.notify();
    }

    getNode(nodeId) {
        if (!this.currentGraph) return null;
        return this.currentGraph.nodes.find(n => n.id === nodeId);
    }

    hasConnection(sourceNodeId, sourceOutput, targetNodeId, targetInput) {
        if (!this.currentGraph) return false;

        const sourceNode = this.getNode(sourceNodeId);
        if (!sourceNode) return false;

        const output = (sourceNode.outputs || []).find(o => o.name === sourceOutput);
        if (!output) return false;

        return (output.connections || []).some(conn =>
            conn.node_id === targetNodeId && conn.input_name === targetInput
        );
    }
}
