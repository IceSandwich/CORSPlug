class CORSPlug {
    private m_sessionID: string = ""
    private m_corsPlugPort: number = 0
    private constructor() { }

    refineRawURL(url: string): string {
        if (url == "") url = "/"

        return `http://127.0.0.1:${this.m_corsPlugPort}/${this.m_sessionID}${url}`
    }

    Get(url: string, headers: Map<string, string> | null = null) {
        var xhr = new XMLHttpRequest();
        xhr.open("GET", this.refineRawURL(url), false)
        if (headers != null) {
            for (let [key, value] of headers) {
                xhr.setRequestHeader(key, value);
            }
        }
        xhr.send()
        return xhr
    }

    Post(url: string, content_type: string, data: any, asyncFunc: ((xhr: XMLHttpRequest) => void) | null  = null, headers: Map<string, string> | null = null) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", this.refineRawURL(url), asyncFunc != null)
        xhr.setRequestHeader('Content-Type', content_type);
        if (headers != null) {
            for (let [key, value] of headers) {
                xhr.setRequestHeader(key, value);
            }
        }
        if (asyncFunc != null) {
            xhr.onload = () => {
                asyncFunc(xhr);
            }
        }
        xhr.send(data)
        if (asyncFunc == null) {
            return xhr
        }
    }

    static New(host: string, corsPlugPort = 11451) {
        var sessionId = ""

        var url = `http://127.0.0.1:${corsPlugPort}/require_permission?host=${host}`

        var xhr = new XMLHttpRequest();
        xhr.open("GET", url, false)
        xhr.send()
        if (xhr.readyState === 4 && xhr.status === 200) {
            var jsonResponse = JSON.parse(xhr.responseText);

            sessionId = jsonResponse["msg"]
            console.log("session id:", sessionId)
            var ret = new CORSPlug()
            ret.m_sessionID = sessionId
            ret.m_corsPlugPort = corsPlugPort
            return ret
        }
        
        console.error(`failed to get session id(${xhr.status}): ${xhr.responseText}`)
        return null
    }
}

