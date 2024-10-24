// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.778
package pages

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import (
	"github.com/a-h/templ"
	templruntime "github.com/a-h/templ/runtime"

	"gitlab.com/hmajid2301/banterbus/internal/views/components"
)

func Index() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>Banter Bus</title><link rel=\"stylesheet\" href=\"/static/css/styles.css\"><link rel=\"icon\" type=\"image/png\" href=\"/static/images/favicon-48x48.png\" sizes=\"48x48\"><link rel=\"icon\" type=\"image/svg+xml\" href=\"/static/images/favicon.svg\"><link rel=\"shortcut icon\" href=\"/static/images/favicon.ico\"><link rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"/static/images/apple-touch-icon.png\"><meta name=\"apple-mobile-web-app-title\" content=\"Banter Bus\"><link rel=\"manifest\" href=\"/static/site.webmanifest\"></head><body><div class=\"w-full min-h-screen text-lg bg-center bg-no-repeat bg-cover bg-mantle bg-background\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = components.Header().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<section class=\"flex flex-col justify-center items-center min-h-screen\"><div class=\"py-10 px-20 rounded-xl bg-surface0\" hx-ext=\"ws\" ws-connect=\"/ws\"><div class=\"flex flex-col justify-center items-center\"><div class=\"flex flex-col items-center space-y-10\"><h1 class=\"text-8xl tracking-tighter text-center text-text font-header text-shadow-custom\">Banter Bus</h1><div id=\"page\" class=\"mt-5 w-full font-main\"><div class=\"flex flex-col my-1\"><div x-data=\"{ action: &#39;&#39; }\"><form ws-send x-bind:hx-vals=\"action\"><div class=\"mb-5\"><label for=\"nickname\" class=\"block mb-2 font-medium text-text2\">Nickname</label> <input type=\"text\" name=\"nickname\" class=\"py-3 px-5 w-full rounded-xl border-1 bg-overlay0 placeholder-surface0 border-text2\" placeholder=\"Enter your nickname\"></div><div class=\"mb-5\"><label for=\"room_code\" class=\"block mb-2 font-medium text-text2\">Room Code</label> <input type=\"text\" name=\"room_code\" class=\"py-3 px-5 w-full rounded-xl border-1 bg-overlay0 placeholder-surface0 border-text2 border-\" placeholder=\"ABC12\"></div><div class=\"flex flex-row mt-12 space-x-4\"><button class=\"flex flex-row justify-center p-3 w-full text-3xl rounded-lg rounded-b-lg text-text font-button shadow-custom-border fill-transparent bg-surface2\" @click=\"action = JSON.stringify({ message_type: &#39;create_room&#39;, game_name: &#39;fibbing_it&#39; })\"><p>Start</p></button> <button class=\"flex flex-row justify-center p-3 w-full text-3xl text-black rounded-lg rounded-b-lg font-button shadow-custom-border fill-transparent bg-text2\" @click=\"action = JSON.stringify({ message_type: &#39;join_lobby&#39; })\"><p>Join</p></button></div></form></div></div></div></div></div></div></section><div id=\"error\"></div><div id=\"spinner\" class=\"grid hidden overflow-x-scroll place-items-center p-6 w-full bg-indigo-500 rounded-lg lg:overflow-visible min-h-[140px]\"><svg class=\"mr-3 -ml-1 w-5 h-5 text-black animate-spin\" xmlns=\"http://www.w3.org/2000/svg\" fill=\"none\" viewBox=\"0 0 24 24\"><circle class=\"opacity-25\" cx=\"12\" cy=\"12\" r=\"10\" stroke=\"currentColor\" stroke-width=\"4\"></circle> <path class=\"opacity-75\" fill=\"currentColor\" d=\"M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z\"></path></svg></div></div></body><script src=\"https://unpkg.com/htmx.org@2.0.2\" integrity=\"sha384-Y7hw+L/jvKeWIRRkqWYfPcvVxHzVzn5REgzbawhxAuQGwX1XWe70vji+VSeHOThJ\" crossorigin=\"anonymous\"></script><script src=\"https://unpkg.com/htmx-ext-ws@2.0.0/ws.js\"></script><script defer src=\"https://cdn.jsdelivr.net/npm/alpinejs@3.14.3/dist/cdn.min.js\"></script><script>\n            htmx.on(\"htmx:wsAfterSend\", (evt) => {\n                document.getElementById('spinner').classList.remove('hidden');\n            });\n            htmx.on(\"htmx:wsAfterMessage\", (evt) => {\n                document.getElementById('spinner').classList.add('hidden');\n            });\n      </script></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
