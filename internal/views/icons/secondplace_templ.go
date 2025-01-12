// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package icons

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import (
	"github.com/a-h/templ"
	templruntime "github.com/a-h/templ/runtime"
)

func SecondPlace(className string) templ.Component {
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
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/views/icons/secondplace.templ`, Line: 1, Col: 0}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><path fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M15.9285 2.62141C16.1336 2.10863 15.8842 1.52666 15.3714 1.32155C14.8586 1.11643 14.2766 1.36585 14.0715 1.87863L13.0715 4.37863C12.8664 4.89141 13.1158 5.47339 13.6286 5.6785C14.1414 5.88361 14.7234 5.6342 14.9285 5.12141L15.9285 2.62141ZM10.4191 1.85611C10.2016 1.34848 9.61371 1.11332 9.10608 1.33087C8.59845 1.54843 8.36329 2.13631 8.58085 2.64394L11.012 8.31668C10.5221 8.38341 10.0478 8.49907 9.59419 8.65858L6.40618 1.82714C6.17262 1.32667 5.57758 1.11029 5.07711 1.34384C4.57664 1.57739 4.36026 2.17244 4.59381 2.67291L7.81621 9.57804C5.96092 10.8911 4.74976 13.0541 4.74976 15.5C4.74976 19.504 7.99569 22.75 11.9998 22.75C16.0038 22.75 19.2498 19.504 19.2498 15.5C19.2498 13.0542 18.0387 10.8914 16.1837 9.5783L19.4062 2.67291C19.6397 2.17244 19.4234 1.57739 18.9229 1.34384C18.4224 1.11029 17.8274 1.32667 17.5938 1.82714L14.4057 8.65873C14.0184 8.52252 13.6161 8.41827 13.2018 8.34915L10.4191 1.85611ZM11.5705 13.9489L11.0494 14.5102C10.7676 14.8138 10.2931 14.8314 9.98951 14.5496C9.68594 14.2678 9.66831 13.7933 9.95012 13.4897L10.497 12.9011C11.4437 11.9437 13.0117 12.058 13.8129 13.1384C14.4358 13.9784 14.3883 15.1442 13.6978 15.93L12.3452 17.25H13.4312C13.8454 17.25 14.1812 17.5857 14.1812 18C14.1812 18.4142 13.8454 18.75 13.4312 18.75H10.4998C10.1946 18.75 9.91987 18.5651 9.80497 18.2823C9.69007 17.9996 9.75793 17.6755 9.97659 17.4626L12.5866 14.9214C12.7949 14.6677 12.8056 14.2983 12.608 14.0319C12.3548 13.6905 11.8694 13.6547 11.5705 13.9489Z\" fill=\"currentColor\"></path></svg>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate