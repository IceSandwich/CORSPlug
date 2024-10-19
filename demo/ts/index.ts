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
        return models
    }

    Generate(request: OllamaGenerateRequest) {
        var xhr = this.m_net.Post("/api/generate", "application/json", JSON.stringify(request));
        var response: OllamaGenerateResponse = JSON.parse(xhr.responseText);
        return response
    }

    static New(corsPlug: CORSPlug) {
        return new Ollama(corsPlug);
    }
}
let ollama: Ollama | null = null;

function init() {
    const select = document.getElementById("models")!;
    // 清空下拉列表（如果需要）
    select.innerHTML = '';

    // 添加选项到下拉列表
    var models = ollama!.ListModels();
    models.forEach(model => {
        const newOption = document.createElement('option');
        newOption.value = model.model;
        newOption.textContent = model.name;
        select.appendChild(newOption);
    })
}

function onClickGetPermission() {
    var corsPlug = CORSPlug.New("127.0.0.1:11434")
    if (corsPlug != null) {
        ollama = Ollama.New(corsPlug);
        init();
    }
}

function onClickGenerate() {
    var prompt = (document.getElementById("prompt") as HTMLInputElement).value!
    var model = (document.getElementById("models") as HTMLInputElement).value!;
    var request: OllamaGenerateRequest = {
        model: model,
        stream: false
    }
}