// API client for backend communication

const API_BASE = '/api';

export async function listImageGraphs() {
    const response = await fetch(`${API_BASE}/imagegraphs`);
    if (!response.ok) {
        throw new Error(`Failed to list image graphs: ${response.statusText}`);
    }
    const data = await response.json();
    return data.imagegraphs;
}

export async function createImageGraph(name) {
    const response = await fetch(`${API_BASE}/imagegraphs`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name }),
    });
    if (!response.ok) {
        throw new Error(`Failed to create image graph: ${response.statusText}`);
    }
    const data = await response.json();
    return data.id;
}

export async function getImageGraph(id) {
    const response = await fetch(`${API_BASE}/imagegraphs/${id}`);
    if (!response.ok) {
        throw new Error(`Failed to get image graph: ${response.statusText}`);
    }
    return response.json();
}

export async function addNode(graphId, nodeType) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/nodes`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            type: nodeType,
            name: nodeType,
            config: "{}",
        }),
    });
    if (!response.ok) {
        throw new Error(`Failed to add node: ${response.statusText}`);
    }
    const data = await response.json();
    return data.node;
}

export async function deleteNode(graphId, nodeId) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/nodes/${nodeId}`, {
        method: 'DELETE',
    });
    if (!response.ok) {
        throw new Error(`Failed to delete node: ${response.statusText}`);
    }
}

export async function connectNodes(graphId, sourceNodeId, sourceOutput, targetNodeId, targetInput) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/connectNodes`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            source_node_id: sourceNodeId,
            source_output: sourceOutput,
            target_node_id: targetNodeId,
            target_input: targetInput,
        }),
    });
    if (!response.ok) {
        throw new Error(`Failed to connect nodes: ${response.statusText}`);
    }
}

export async function disconnectNodes(graphId, sourceNodeId, sourceOutput, targetNodeId, targetInput) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/disconnectNodes`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            source_node_id: sourceNodeId,
            source_output: sourceOutput,
            target_node_id: targetNodeId,
            target_input: targetInput,
        }),
    });
    if (!response.ok) {
        throw new Error(`Failed to disconnect nodes: ${response.statusText}`);
    }
}

export async function setNodeConfig(graphId, nodeId, config) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/nodes/${nodeId}/config`, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ config }),
    });
    if (!response.ok) {
        throw new Error(`Failed to set node config: ${response.statusText}`);
    }
}

export async function setNodeOutputImage(graphId, nodeId, outputName, imageData) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/nodes/${nodeId}/outputs/${outputName}`, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ image: imageData }),
    });
    if (!response.ok) {
        throw new Error(`Failed to set node output image: ${response.statusText}`);
    }
}
