export interface ChannelCatalogEntry {
	code: string
	name: string
	level: 'A' | 'B' | 'C'
	region: string
	status: '活跃' | '暂停' | '待审核'
	commissionRate: number
	settlement: '月结' | '半月结' | '季结'
	contact: string
	gmv30d: number
}

export const channelCatalog: ChannelCatalogEntry[] = [
	{
		code: 'official',
		name: '官方商城',
		level: 'A',
		region: '全国',
		status: '活跃',
		commissionRate: 0.05,
		settlement: '月结',
		contact: '商城团队',
		gmv30d: 680000,
	},
	{
		code: 'mini_app',
		name: '直营小程序',
		level: 'A',
		region: '全国',
		status: '活跃',
		commissionRate: 0.04,
		settlement: '半月结',
		contact: '小程序运营',
		gmv30d: 520000,
	},
	{
		code: 'market_a',
		name: '第三方市场 A',
		level: 'B',
		region: '华东',
		status: '暂停',
		commissionRate: 0.06,
		settlement: '月结',
		contact: '渠道 A 团队',
		gmv30d: 120000,
	},
	{
		code: 'market_b',
		name: '第三方市场 B',
		level: 'C',
		region: '华南',
		status: '待审核',
		commissionRate: 0.05,
		settlement: '季结',
		contact: '渠道 B 团队',
		gmv30d: 80000,
	},
]

export const channelSelectOptions = channelCatalog.map((channel) => ({
	label: channel.name,
	value: channel.code,
}))
