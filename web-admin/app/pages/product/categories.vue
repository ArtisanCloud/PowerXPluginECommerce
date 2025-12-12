<template>
	<div class="px-6 py-10 text-white">
		<div class="mx-auto flex max-w-5xl flex-col gap-6">
			<header class="space-y-2">
				<p class="text-xs uppercase tracking-[0.2em] text-primary-200">商品中心</p>
				<h1 class="text-3xl font-semibold">类目管理</h1>
				<p class="text-sm text-white/70">类目树、属性模板绑定与推荐位配置仍在设计中，当前页面仅作为占位，避免导航访问报错。</p>
			</header>

			<UCard class="border border-white/10 bg-white/5 shadow-lg ring-1 ring-white/5 backdrop-blur">
				<div class="space-y-6">
					<section class="grid gap-4 md:grid-cols-2">
						<div class="rounded-xl border border-white/10 bg-white/5 p-4">
							<h2 class="text-base font-semibold">类目树</h2>
							<p class="text-sm text-white/60">后续将支持拖拽排序、批量调整以及类目 SEO 字段。</p>
						</div>
						<div class="rounded-xl border border-white/10 bg-white/5 p-4">
							<h2 class="text-base font-semibold">属性模板 & 运营配置</h2>
							<p class="text-sm text-white/60">包括属性模板绑定、推荐位、列表展示策略等高级设置。</p>
						</div>
					</section>

					<div class="flex flex-wrap items-center justify-between gap-3">
						<div class="text-sm text-white/60">类目功能尚未开放，设计稿确认后即会上线。</div>
						<UButton color="primary" disabled variant="soft">敬请期待</UButton>
					</div>
				</div>
			</UCard>
		</div>
	</div>
</template>
