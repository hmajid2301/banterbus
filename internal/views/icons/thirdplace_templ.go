// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package icons

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import (
	"github.com/a-h/templ"
	templruntime "github.com/a-h/templ/runtime"
)

func ThirdPlace(className string) templ.Component {
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
		var templ_7745c5c3_Var2 = []any{className}
		templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var2...)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\" width=\"24\" height=\"24\" fill=\"none\" class=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var2).String())
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/views/icons/thirdplace.templ`, Line: 1, Col: 0}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><path fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M15.9285 2.62141C16.1336 2.10863 15.8842 1.52666 15.3714 1.32155C14.8586 1.11643 14.2766 1.36585 14.0715 1.87863L13.0715 4.37863C12.8664 4.89141 13.1158 5.47339 13.6286 5.6785C14.1414 5.88361 14.7234 5.6342 14.9285 5.12141L15.9285 2.62141ZM10.4191 1.85611C10.2016 1.34848 9.61371 1.11332 9.10608 1.33087C8.59845 1.54843 8.36329 2.13631 8.58085 2.64394L11.012 8.31668C10.5221 8.38341 10.0478 8.49907 9.59419 8.65858L6.40618 1.82714C6.17262 1.32667 5.57758 1.11029 5.07711 1.34384C4.57664 1.57739 4.36026 2.17244 4.59381 2.67291L7.81621 9.57804C5.96092 10.8911 4.74976 13.0541 4.74976 15.5C4.74976 19.504 7.99569 22.75 11.9998 22.75C16.0038 22.75 19.2498 19.504 19.2498 15.5C19.2498 13.0542 18.0387 10.8914 16.1837 9.5783L19.4062 2.67291C19.6397 2.17244 19.4234 1.57739 18.9229 1.34384C18.4224 1.11029 17.8274 1.32667 17.5938 1.82714L14.4057 8.65873C14.0184 8.52252 13.6161 8.41827 13.2018 8.34915L10.4191 1.85611ZM11.5278 13.8724C11.4192 13.9473 11.2974 14.0821 11.1707 14.3354C10.9855 14.7058 10.535 14.856 10.1645 14.6708C9.79403 14.4855 9.64386 14.035 9.8291 13.6645C10.029 13.2648 10.2964 12.8996 10.6758 12.6378C11.0639 12.37 11.5115 12.25 11.9999 12.25C12.5092 12.25 13.0466 12.4064 13.474 12.7275C13.9158 13.0594 14.2499 13.5812 14.2499 14.25C14.2499 14.7229 14.0857 15.1576 13.8113 15.5C14.0857 15.8424 14.2499 16.277 14.2499 16.75C14.2499 17.4187 13.9158 17.9405 13.474 18.2724C13.0466 18.5935 12.5092 18.75 11.9999 18.75C11.5115 18.75 11.0639 18.6299 10.6758 18.3621C10.2964 18.1003 10.029 17.7351 9.8291 17.3354C9.64386 16.9649 9.79403 16.5144 10.1645 16.3291C10.535 16.1439 10.9855 16.2941 11.1707 16.6645C11.2974 16.9178 11.4192 17.0526 11.5278 17.1275C11.6278 17.1965 11.7669 17.25 11.9999 17.25C12.2202 17.25 12.4328 17.1785 12.573 17.0732C12.6988 16.9786 12.7499 16.8754 12.7499 16.75C12.7499 16.4738 12.5261 16.25 12.2499 16.25C11.8357 16.25 11.4999 15.9142 11.4999 15.5C11.4999 15.0857 11.8357 14.75 12.2499 14.75C12.5261 14.75 12.7499 14.5261 12.7499 14.25C12.7499 14.1245 12.6988 14.0213 12.573 13.9268C12.4328 13.8214 12.2202 13.75 11.9999 13.75C11.7669 13.75 11.6278 13.8034 11.5278 13.8724Z\" fill=\"currentColor\"></path></svg>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
