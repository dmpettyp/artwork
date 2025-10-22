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

export async function addNode(graphId, nodeType, nodeName, config) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/nodes`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            type: nodeType,
            name: nodeName,
            config: config, // config is now sent as an object, not a string
        }),
    });
    if (!response.ok) {
        throw new Error(`Failed to add node: ${response.statusText}`);
    }
    const data = await response.json();
    return data.id;
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
            from_node_id: sourceNodeId,
            output_name: sourceOutput,
            to_node_id: targetNodeId,
            input_name: targetInput,
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
            from_node_id: sourceNodeId,
            output_name: sourceOutput,
            to_node_id: targetNodeId,
            input_name: targetInput,
        }),
    });
    if (!response.ok) {
        throw new Error(`Failed to disconnect nodes: ${response.statusText}`);
    }
}

export async function updateNode(graphId, nodeId, name, config) {
    const body = {};
    if (name !== undefined && name !== null) {
        body.name = name;
    }
    if (config !== undefined && config !== null) {
        body.config = config;
    }

    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/nodes/${nodeId}`, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
    });
    if (!response.ok) {
        throw new Error(`Failed to update node: ${response.statusText}`);
    }
}

export async function uploadNodeOutputImage(graphId, nodeId, outputName, imageFile) {
    const formData = new FormData();
    formData.append('image', imageFile);

    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/nodes/${nodeId}/outputs/${outputName}`, {
        method: 'PUT',
        body: formData,
    });
    if (!response.ok) {
        throw new Error(`Failed to upload node output image: ${response.statusText}`);
    }
    const data = await response.json();
    return data.image_id;
}

// Layout API functions

export async function getLayout(graphId) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/layout`);
    if (!response.ok) {
        throw new Error(`Failed to get layout: ${response.statusText}`);
    }
    return response.json();
}

export async function updateLayout(graphId, nodePositions) {
    // Convert nodePositions Map to array format
    const nodePositionsArray = Array.from(nodePositions.entries()).map(([nodeId, pos]) => ({
        node_id: nodeId,
        x: pos.x,
        y: pos.y,
    }));

    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/layout`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            node_positions: nodePositionsArray,
        }),
    });
    if (!response.ok) {
        throw new Error(`Failed to update layout: ${response.statusText}`);
    }
}

// Viewport API functions

export async function getViewport(graphId) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/viewport`);
    if (!response.ok) {
        throw new Error(`Failed to get viewport: ${response.statusText}`);
    }
    return response.json();
}

export async function updateViewport(graphId, viewport) {
    const response = await fetch(`${API_BASE}/imagegraphs/${graphId}/viewport`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            zoom: viewport.zoom,
            pan_x: viewport.panX,
            pan_y: viewport.panY,
        }),
    });
    if (!response.ok) {
        throw new Error(`Failed to update viewport: ${response.statusText}`);
    }
}
