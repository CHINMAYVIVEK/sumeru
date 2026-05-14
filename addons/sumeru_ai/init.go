package sumeru_ai

import (
	"context"
	"html/template"
	"log"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

func init() {
	log.Println("Initializing Sumeru AI Addon...")
	
	// Register ORM Search Interceptor
	orm.RegisterSearchInterceptor(func(ctx context.Context, model string, domain [][]interface{}) ([][]interface{}, error) {
		log.Printf("[AI INTERCEPT] Search on model: %s, Domain: %v", model, domain)
		// Here we would call LLM to translate natural language if needed.
		// For now, we just pass through.
		return domain, nil
	})

	// Register Shell Hook for AI Assistant
	render.RegisterShellHook(func(ctx context.Context, vr *render.ViewRecordData, ro bool) template.HTML {
		return template.HTML(`
			<div id="sumeru-ai-assistant" class="fixed bottom-6 right-6 z-[9999]">
				<button class="w-14 h-14 bg-indigo-600 text-white rounded-full shadow-2xl flex items-center justify-center hover:bg-indigo-700 transition-all group">
					<svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"></path></svg>
					<span class="absolute right-full mr-4 bg-slate-900 text-white text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 whitespace-nowrap transition-opacity pointer-events-none">AI Assistant</span>
				</button>
			</div>
		`)
	})

	// Register Notebook Hook for core.partner
	render.RegisterNotebookHook("core.partner", "ai insights", func(ctx context.Context, vr *render.ViewRecordData, ro bool) template.HTML {
		return template.HTML(`
			<div class="p-6 bg-indigo-50/50 rounded-xl border border-indigo-100">
				<h3 class="text-indigo-900 font-semibold mb-4 flex items-center gap-2">
					<svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
					AI Relationship Insights
				</h3>
				<p class="text-sm text-indigo-800/80 leading-relaxed">
					This partner shows high potential for churn. AI suggests reaching out with a personalized discount code.
				</p>
			</div>
		`)
	})
}
