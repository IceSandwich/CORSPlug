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
        stream: false,
        prompt: prompt,
    };
    ollama?.Generate(request, (rep) => {
        document.getElementById("Response")!.textContent = rep.response;
    })
}