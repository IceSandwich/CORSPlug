interface OllamaTagsResponse {
    name: string
    model: string
}

interface ModelOptions {
    seed?: number
    num_keep?: number
    num_predict?: number
    top_k?: number
    top_p?: number 
    min_p?: number 
    tfs_z?: number 
    typical_p?: number
    repeat_last_n?: number
    temperature?: number
    repeat_penalty?: number
    presence_penalty?: number
    frequency_penalty?: number
    mirostat?: number
    mirostat_tau?: number
    mirostat_eta?: number
    penalize_newline?: boolean
    stop?: string[]
    numa?: boolean
    num_ctx?: number
    num_batch?: number
    num_gpu?: number
    main_gpu?: number
    low_vram?: boolean
    f16_kv?: boolean
    vocab_only?: boolean
    use_mmap?: boolean
    use_mlock?: boolean
    num_thread?: number
}

interface OllamaGenerateRequest {
    model: string
    prompt: string
    stream: boolean
    raw?: boolean
    options?: ModelOptions
    format?: string
    // images in base64 encoding
    images?: string[]
}

interface OllamaGenerateResponse {
    model: string
    response: string
    done: boolean
}

class Ollama {
    private m_net: CORSPlug
    private constructor(net: CORSPlug) {
        this.m_net = net;
    }

    ListModels() {
        var xhr = this.m_net.Get("/api/tags");
        var models: OllamaTagsResponse[] = JSON.parse(xhr.responseText)["models"];
        return models;
    }

    Generate(request: OllamaGenerateRequest, callback: (response: OllamaGenerateResponse) => void) {
        this.m_net.Post("/api/generate", 'application/json;charset=UTF-8', JSON.stringify(request), (xhr) => {
            callback(JSON.parse(xhr.responseText));
        });
    }

    static New(corsPlug: CORSPlug) {
        return new Ollama(corsPlug);
    }
}