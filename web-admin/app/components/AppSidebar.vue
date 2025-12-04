<template>
  <aside
    class="w-64 min-w-64 max-w-64 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 min-h-screen flex-shrink-0 overflow-y-auto"
  >
    <nav class="p-4 space-y-4">
      <!-- 经营总览 -->
      <div>
        <UButton
          to="/dashboard"
          variant="ghost"
          color="neutral"
          class="w-full justify-start"
          :class="{
            'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
              $route.path === '/dashboard',
          }"
        >
          <UIcon name="i-heroicons-chart-bar" class="w-4 h-4 mr-3"/>
          交易仪表盘
        </UButton>
      </div>

      <!-- 客户运营 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          客户运营
        </div>
        <div class="space-y-1">
          <UButton
            to="/customer"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/customer',
            }"
          >
            <UIcon name="i-heroicons-users" class="w-4 h-4 mr-3"/>
            客户
          </UButton>

          <UButton
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            @click="toggleMembershipTiers"
          >
            <UIcon name="i-heroicons-star" class="w-4 h-4 mr-3"/>
            会员等级
            <UIcon
              :name="
                showMembershipTiers
                  ? 'i-heroicons-chevron-down'
                  : 'i-heroicons-chevron-right'
              "
              class="w-4 h-4 ml-auto"
            />
          </UButton>

          <div v-show="showMembershipTiers" class="ml-6 mt-1 space-y-1">
            <UButton
              to="/customer/membership/tiers"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/membership/tiers',
              }"
            >
              <UIcon name="i-heroicons-list-bullet" class="w-3 h-3 mr-2"/>
              等级管理
            </UButton>
            <UButton
              to="/customer/membership/benefits"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/membership/benefits',
              }"
            >
              <UIcon name="i-heroicons-gift" class="w-3 h-3 mr-2"/>
              等级权益
            </UButton>
            <UButton
              to="/customer/membership/rules"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/membership/rules',
              }"
            >
              <UIcon name="i-heroicons-cog-6-tooth" class="w-3 h-3 mr-2"/>
              等级规则
            </UButton>
            <UButton
              to="/customer/membership/analytics"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/membership/analytics',
              }"
            >
              <UIcon name="i-heroicons-chart-bar" class="w-3 h-3 mr-2"/>
              客户分布分析
            </UButton>
          </div>

          <UButton
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            @click="toggleMembershipPointsAndGrowth"
          >
            <UIcon name="i-heroicons-gift" class="w-4 h-4 mr-3"/>
            积分与成长值
            <UIcon
              :name="
                showMembershipPointsAndGrowth
                  ? 'i-heroicons-chevron-down'
                  : 'i-heroicons-chevron-right'
              "
              class="w-4 h-4 ml-auto"
            />
          </UButton>

          <div v-show="showMembershipPointsAndGrowth" class="ml-6 mt-1 space-y-1">
            <!-- 积分管理子菜单 -->
            <UButton
              to="/customer/membership/points"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/membership/points' || $route.path === '/customer/membership/points/accounts',
              }"
            >
              <UIcon name="i-heroicons-cog-6-tooth" class="w-3 h-3 mr-2"/>
              积分规则
            </UButton>

            <UButton
              to="/customer/membership/points/accounts"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/membership/points/accounts',
              }"
            >
              <UIcon name="i-heroicons-wallet" class="w-3 h-3 mr-2"/>
              积分账户
            </UButton>


            <!-- 成长值管理子菜单 -->
            <UButton
              to="/customer/membership/growth-value/promotion-rules"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/membership/growth-value/promotion-rules',
              }"
            >
              <UIcon name="i-heroicons-cog-6-tooth" class="w-3 h-3 mr-2"/>
              晋升规则
            </UButton>

            <UButton
              to="/customer/membership/growth-value"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/membership/growth-value',
              }"
            >
              <UIcon name="i-heroicons-wallet" class="w-3 h-3 mr-2"/>
              成长值账户
            </UButton>

            <UButton
              to="/customer/membership/points/mall"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/membership/points/mall',
              }"
            >
              <UIcon name="i-heroicons-shopping-bag" class="w-3 h-3 mr-2"/>
              积分商城
            </UButton>

          </div>

          <UButton
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            @click="toggleAffiliates"
          >
            <UIcon name="i-heroicons-user-plus" class="w-4 h-4 mr-3"/>
            推荐/分销
            <UIcon
              :name="
                showAffiliates
                  ? 'i-heroicons-chevron-down'
                  : 'i-heroicons-chevron-right'
              "
              class="w-4 h-4 ml-auto"
            />
          </UButton>

          <div v-show="showAffiliates" class="ml-6 mt-1 space-y-1">
            <UButton
              to="/customer/affiliates"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/affiliates',
              }"
            >
              <UIcon name="i-heroicons-link" class="w-3 h-3 mr-2"/>
              邀请记录
            </UButton>
            <UButton
              to="/customer/affiliates/distributors"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/affiliates/distributors',
              }"
            >
              <UIcon name="i-heroicons-user-group" class="w-3 h-3 mr-2"/>
              分销员管理
            </UButton>
            <UButton
              to="/customer/affiliates/settlements"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/affiliates/settlements',
              }"
            >
              <UIcon name="i-heroicons-currency-dollar" class="w-3 h-3 mr-2"/>
              佣金结算
            </UButton>
            <UButton
              to="/customer/affiliates/settlements/review"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/affiliates/settlements/review',
              }"
            >
              <UIcon name="i-heroicons-document-check" class="w-3 h-3 mr-2"/>
              提现申请审核
            </UButton>
            <UButton
              to="/customer/affiliates/rules"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/affiliates/rules',
              }"
            >
              <UIcon name="i-heroicons-cog-6-tooth" class="w-3 h-3 mr-2"/>
              规则配置
            </UButton>
          </div>

          <UButton
            variant="ghost"
            color="neutral"
            class="w-full justify-start text-gray-500"
            @click="toggleReturns"
          >
            <UIcon name="i-heroicons-arrow-uturn-left" class="w-4 h-4 mr-3" />
            客户自助退换货门户
            <UIcon
              :name="
                showReturns
                  ? 'i-heroicons-chevron-down'
                  : 'i-heroicons-chevron-right'
              "
              class="w-4 h-4 ml-auto"
            />
          </UButton>

          <div v-show="showReturns" class="ml-6 mt-1 space-y-1">
            <UButton
              to="/customer/returns/config"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/returns/config',
              }"
            >
              <UIcon name="i-heroicons-cog-6-tooth" class="w-3 h-3 mr-2"/>
              RMA 入口配置
            </UButton>
            <UButton
              to="/customer/returns/management"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/returns/management',
              }"
            >
              <UIcon name="i-heroicons-clipboard-document-list" class="w-3 h-3 mr-2"/>
              退换货申请管理
            </UButton>
            <UButton
              to="/customer/returns/process"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/returns/process',
              }"
            >
              <UIcon name="i-heroicons-arrow-path" class="w-3 h-3 mr-2"/>
              RMA 流程配置
            </UButton>
            <UButton
              to="/customer/returns/notifications"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/customer/returns/notifications',
              }"
            >
              <UIcon name="i-heroicons-bell" class="w-3 h-3 mr-2"/>
              客户通知
            </UButton>
          </div>
        </div>
      </div>

      <!-- 商品中心 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          商品中心
        </div>
        <div class="space-y-1">
          <UButton
            to="/product/spus"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/product/spus',
            }"
          >
            <UIcon name="i-heroicons-cube" class="w-4 h-4 mr-3"/>
            商品（SPU）
          </UButton>

          <UButton
            to="/product/skus"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/product/skus',
            }"
          >
            <UIcon name="i-heroicons-squares-2x2" class="w-4 h-4 mr-3"/>
            SKU & 变体
          </UButton>

          <UButton
            to="/product/categories"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/product/categories',
            }"
          >
            <UIcon name="i-heroicons-tag" class="w-4 h-4 mr-3"/>
            类目
          </UButton>

          <UButton
            to="/product/brands"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/product/brands',
            }"
          >
            <UIcon
              name="i-heroicons-building-storefront"
              class="w-4 h-4 mr-3"
            />
            品牌
          </UButton>

          <NuxtLink 
            to="/product/attributes" 
            class="block w-full"
            @click="toggleAttributeAndSpec"
          >
            <UButton
              variant="ghost"
              color="neutral"
              class="w-full justify-start"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path.startsWith('/product/attributes'),
              }"
            >
              <UIcon
                name="i-heroicons-adjustments-horizontal"
                class="w-4 h-4 mr-3"
              />
              属性和规格
              <UIcon
                :name="
                  showAttributeAndSpec
                    ? 'i-heroicons-chevron-down'
                    : 'i-heroicons-chevron-right'
                "
                class="w-4 h-4 ml-auto"
              />
            </UButton>
          </NuxtLink>

          <div v-show="showAttributeAndSpec" class="ml-6 mt-1 space-y-1">
            <UButton
              to="/product/attributes"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/product/attributes',
              }"
            >
              <UIcon name="i-heroicons-tag" class="w-3 h-3 mr-2"/>
              属性
            </UButton>
            
            <UButton
              to="/product/specifications"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/product/specifications',
              }"
            >
              <UIcon name="i-heroicons-ruler" class="w-3 h-3 mr-2"/>
              规格
            </UButton>
            
            <UButton
              to="/product/attribute-groups"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/product/attribute-groups',
              }"
            >
              <UIcon name="i-heroicons-folder" class="w-3 h-3 mr-2"/>
              属性组
            </UButton>
            
            <UButton
              to="/product/category-templates"
              variant="ghost"
              color="neutral"
              size="sm"
              class="w-full justify-start text-sm"
              :class="{
                'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                  $route.path === '/product/category-templates',
              }"
            >
              <UIcon name="i-heroicons-template" class="w-3 h-3 mr-2"/>
              类目模板
            </UButton>
          </div>

          <UButton
            to="/product/media"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/product/media',
            }"
          >
            <UIcon name="i-heroicons-photo" class="w-4 h-4 mr-3"/>
            媒体与素材
          </UButton>

          <UButton
            to="/product/suppliers"
            variant="ghost"
            color="neutral"
            class="w-full justify-start text-gray-500"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/product/suppliers',
            }"
          >
            <UIcon name="i-heroicons-building-office" class="w-4 h-4 mr-3"/>
            供应商
          </UButton>
        </div>
      </div>

      <!-- 定价中心 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          定价中心
        </div>
        <div class="space-y-1">
          <UButton
            to="/pricing"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/pricing',
            }"
          >
            <UIcon name="i-heroicons-currency-dollar" class="w-4 h-4 mr-3"/>
            价目表
          </UButton>

          <UButton
            to="/pricing/tiers"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/pricing/tiers',
            }"
          >
            <UIcon name="i-heroicons-bars-3-bottom-left" class="w-4 h-4 mr-3"/>
            阶梯价/客户专属价
          </UButton>

          <UButton
            to="/pricing/agreements"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/pricing/agreements',
            }"
          >
            <UIcon name="i-heroicons-document-text" class="w-4 h-4 mr-3"/>
            合同价/协议价（B2B）
          </UButton>

          <UButton
            to="/pricing/promotions"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/pricing/promotions',
            }"
          >
            <UIcon name="i-heroicons-sparkles" class="w-4 h-4 mr-3"/>
            促销规则
          </UButton>

          <UButton
            to="/pricing/coupons"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/pricing/coupons',
            }"
          >
            <UIcon name="i-heroicons-ticket" class="w-4 h-4 mr-3"/>
            优惠券
          </UButton>

          <UButton
            to="/pricing/giftcards"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/pricing/giftcards',
            }"
          >
            <UIcon name="i-heroicons-gift" class="w-4 h-4 mr-3"/>
            礼品卡
          </UButton>
        </div>
      </div>

      <!-- 营销增长 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          营销增长
        </div>
        <div class="space-y-1">
          <!-- 订单管理 - 带子菜单 -->
          <div>
            <UButton
              variant="ghost"
              color="neutral"
              class="w-full justify-start"
              @click="toggleOrderManagement"
            >
              <UIcon
                name="i-heroicons-clipboard-document-list"
                class="w-4 h-4 mr-3"
              />
              订单管理
              <UIcon
                :name="
                  showOrderManagement
                    ? 'i-heroicons-chevron-down'
                    : 'i-heroicons-chevron-right'
                "
                class="w-4 h-4 ml-auto"
              />
            </UButton>

            <div v-show="showOrderManagement" class="ml-6 mt-1 space-y-1">
              <UButton
                to="/market/orders"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/orders',
                }"
              >
                <UIcon name="i-heroicons-list-bullet" class="w-3 h-3 mr-2"/>
                订单（全部）
              </UButton>

              <UButton
                to="/market/orders/pending-payment"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/orders/pending-payment',
                }"
              >
                <UIcon name="i-heroicons-clock" class="w-3 h-3 mr-2"/>
                待支付
              </UButton>

              <UButton
                to="/market/orders/to-ship"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/orders/to-ship',
                }"
              >
                <UIcon name="i-heroicons-truck" class="w-3 h-3 mr-2"/>
                待发货
              </UButton>

              <UButton
                to="/market/orders/shipped"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/orders/shipped',
                }"
              >
                <UIcon name="i-heroicons-check-circle" class="w-3 h-3 mr-2"/>
                已发货
              </UButton>
            </div>
          </div>

          <UButton
            to="/market/payment"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/market/payment',
            }"
          >
            <UIcon name="i-heroicons-credit-card" class="w-4 h-4 mr-3"/>
            支付单（交易流水）
          </UButton>

          <!-- 营销互动 - 带子菜单 -->
          <div>
            <UButton
              variant="ghost"
              color="neutral"
              class="w-full justify-start"
              @click="toggleMarketingInteraction"
            >
              <UIcon name="i-heroicons-megaphone" class="w-4 h-4 mr-3"/>
              营销互动
              <UIcon
                :name="
                  showMarketingInteraction
                    ? 'i-heroicons-chevron-down'
                    : 'i-heroicons-chevron-right'
                "
                class="w-4 h-4 ml-auto"
              />
            </UButton>

            <div v-show="showMarketingInteraction" class="ml-6 mt-1 space-y-1">
              <UButton
                to="/market/marketing"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/marketing',
                }"
              >
                <UIcon name="i-heroicons-speaker-wave" class="w-3 h-3 mr-2"/>
                活动中心/Campaign
              </UButton>

              <UButton
                to="/market/personalization"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm text-gray-500"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/personalization',
                }"
              >
                <UIcon name="i-heroicons-user-circle" class="w-3 h-3 mr-2"/>
                个性化推荐
              </UButton>

              <UButton
                to="/market/cms/banners"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm text-gray-500"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/cms/banners',
                }"
              >
                <UIcon name="i-heroicons-photo" class="w-3 h-3 mr-2"/>
                内容位（轮播）
              </UButton>
            </div>
          </div>

          <!-- 售后与纠纷 - 带子菜单 -->
          <div>
            <UButton
              variant="ghost"
              color="neutral"
              class="w-full justify-start"
              @click="toggleAfterSales"
            >
              <UIcon name="i-heroicons-shield-check" class="w-4 h-4 mr-3"/>
              售后与纠纷
              <UIcon
                :name="
                  showAfterSales
                    ? 'i-heroicons-chevron-down'
                    : 'i-heroicons-chevron-right'
                "
                class="w-4 h-4 ml-auto"
              />
            </UButton>

            <div v-show="showAfterSales" class="ml-6 mt-1 space-y-1">
              <UButton
                to="/market/after-sales"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/after-sales',
                }"
              >
                <UIcon
                  name="i-heroicons-arrow-uturn-left"
                  class="w-3 h-3 mr-2"
                />
                售后单（退换）
              </UButton>

              <UButton
                to="/market/refunds"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/refunds',
                }"
              >
                <UIcon name="i-heroicons-banknotes" class="w-3 h-3 mr-2"/>
                退款
              </UButton>

              <UButton
                to="/market/disputes"
                variant="ghost"
                color="neutral"
                size="sm"
                class="w-full justify-start text-sm"
                :class="{
                  'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                    $route.path === '/market/disputes',
                }"
              >
                <UIcon
                  name="i-heroicons-exclamation-triangle"
                  class="w-3 h-3 mr-2"
                />
                纠纷
              </UButton>
            </div>
          </div>
        </div>
      </div>

      <!-- 库存与仓储 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          库存与仓储
        </div>
        <div class="space-y-1">
          <UButton
            to="/inventory/warehouses"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/inventory/warehouses',
            }"
          >
            <UIcon name="i-heroicons-building-office-2" class="w-4 h-4 mr-3"/>
            仓库
          </UButton>

          <UButton
            to="/inventory/stock"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/inventory/stock',
            }"
          >
            <UIcon name="i-heroicons-archive-box" class="w-4 h-4 mr-3"/>
            库存总览
          </UButton>

          <UButton
            to="/inventory/replenishment"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/inventory/replenishment',
            }"
          >
            <UIcon name="i-heroicons-arrow-up-tray" class="w-4 h-4 mr-3"/>
            补货与安全库存
          </UButton>

          <UButton
            to="/inventory/stocktake"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/inventory/stocktake',
            }"
          >
            <UIcon
              name="i-heroicons-clipboard-document-check"
              class="w-4 h-4 mr-3"
            />
            盘点
          </UButton>

          <UButton
            to="/inventory/transfers"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/inventory/transfers',
            }"
          >
            <UIcon name="i-heroicons-arrow-right-circle" class="w-4 h-4 mr-3"/>
            库存调拨
          </UButton>
        </div>
      </div>

      <!-- 履约与物流 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          履约与物流
        </div>
        <div class="space-y-1">
          <UButton
            to="/shipping/carriers"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/shipping/carriers',
            }"
          >
            <UIcon name="i-heroicons-truck" class="w-4 h-4 mr-3"/>
            物流承运商
          </UButton>

          <UButton
            to="/shipping/templates"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/shipping/templates',
            }"
          >
            <UIcon name="i-heroicons-document-duplicate" class="w-4 h-4 mr-3"/>
            运费模板
          </UButton>

          <UButton
            to="/shipping/waybills"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/shipping/waybills',
            }"
          >
            <UIcon name="i-heroicons-map" class="w-4 h-4 mr-3"/>
            运单与轨迹
          </UButton>
        </div>
      </div>

      <!-- 渠道与上架 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          渠道与上架
        </div>
        <div class="space-y-1">
          <UButton
            to="/channels"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/channels',
            }"
          >
            <UIcon name="i-heroicons-globe-alt" class="w-4 h-4 mr-3"/>
            店铺/渠道
          </UButton>

          <UButton
            to="/channels/publishing"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/channels/publishing',
            }"
          >
            <UIcon name="i-heroicons-cloud-arrow-up" class="w-4 h-4 mr-3"/>
            上架发布
          </UButton>

          <UButton
            to="/channels/sku-mapping"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/channels/sku-mapping',
            }"
          >
            <UIcon name="i-heroicons-link" class="w-4 h-4 mr-3"/>
            SKU 映射
          </UButton>
        </div>
      </div>

      <!-- 财税与结算 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          财税与结算
        </div>
        <div class="space-y-1">
          <UButton
            to="/payments/providers"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/payments/providers',
            }"
          >
            <UIcon name="i-heroicons-credit-card" class="w-4 h-4 mr-3"/>
            支付渠道
          </UButton>

          <UButton
            to="/settlements"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/settlements',
            }"
          >
            <UIcon name="i-heroicons-calculator" class="w-4 h-4 mr-3"/>
            结算与分账
          </UButton>

          <UButton
            to="/tax"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/tax',
            }"
          >
            <UIcon name="i-heroicons-receipt-percent" class="w-4 h-4 mr-3"/>
            税率与税则
          </UButton>
        </div>
      </div>

      <!-- 数据与增长分析 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          数据与增长分析
        </div>
        <div class="space-y-1">
          <UButton
            to="/reports/sales"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/reports/sales',
            }"
          >
            <UIcon name="i-heroicons-chart-bar" class="w-4 h-4 mr-3"/>
            销售报表
          </UButton>

          <UButton
            to="/reports/products"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/reports/products',
            }"
          >
            <UIcon name="i-heroicons-chart-pie" class="w-4 h-4 mr-3"/>
            商品报表
          </UButton>

          <UButton
            to="/reports/conversion"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/reports/conversion',
            }"
          >
            <UIcon name="i-heroicons-funnel" class="w-4 h-4 mr-3"/>
            转化漏斗
          </UButton>

          <UButton
            to="/audit"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/audit',
            }"
          >
            <UIcon
              name="i-heroicons-document-magnifying-glass"
              class="w-4 h-4 mr-3"
            />
            审计日志
          </UButton>
        </div>
      </div>

      <!-- 系统设置 -->
      <div>
        <div
          class="px-3 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
        >
          系统设置
        </div>
        <div class="space-y-1">
          <UButton
            to="/settings/orders"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/settings/orders',
            }"
          >
            <UIcon name="i-heroicons-cog-6-tooth" class="w-4 h-4 mr-3"/>
            订单与流程配置
          </UButton>

          <UButton
            to="/settings/currency"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/settings/currency',
            }"
          >
            <UIcon name="i-heroicons-currency-dollar" class="w-4 h-4 mr-3"/>
            货币与小数精度
          </UButton>

          <UButton
            to="/settings/roles"
            variant="ghost"
            color="neutral"
            class="w-full justify-start"
            :class="{
              'bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400':
                $route.path === '/settings/roles',
            }"
          >
            <UIcon name="i-heroicons-user-group" class="w-4 h-4 mr-3"/>
            权限与角色
          </UButton>
        </div>
      </div>
    </nav>
  </aside>
</template>

<script setup lang="ts">
const {t} = useI18n();

// 控制各个子菜单的展开状态
const showOrderManagement = ref(false);
const showMarketingInteraction = ref(false);
const showAfterSales = ref(false);
const showMembershipTiers = ref(false);
const showMembershipPointsAndGrowth = ref(false);
const showAffiliates = ref(false);
const showReturns = ref(false);
const showAttributeAndSpec = ref(false);

// 切换订单管理子菜单
const toggleOrderManagement = () => {
  showOrderManagement.value = !showOrderManagement.value;
};

// 切换营销互动子菜单
const toggleMarketingInteraction = () => {
  showMarketingInteraction.value = !showMarketingInteraction.value;
};

// 切换售后与纠纷子菜单
const toggleAfterSales = () => {
  showAfterSales.value = !showAfterSales.value;
};

// 切换会员等级子菜单
const toggleMembershipTiers = () => {
  showMembershipTiers.value = !showMembershipTiers.value;
};

// 切换积分与成长值子菜单
const toggleMembershipPointsAndGrowth = () => {
  showMembershipPointsAndGrowth.value = !showMembershipPointsAndGrowth.value;
};

// 切换属性和规格子菜单
const toggleAttributeAndSpec = () => {
  showAttributeAndSpec.value = !showAttributeAndSpec.value;
};

// 切换推荐/分销子菜单
const toggleAffiliates = () => {
  showAffiliates.value = !showAffiliates.value;
};

// 切换客户自助退换货门户子菜单
const toggleReturns = () => {
  showReturns.value = !showReturns.value;
};

// 监听路由变化，自动展开相应的子菜单
const route = useRoute();
watch(
  () => route.path,
  (newPath) => {
    // 订单管理相关路由
    if (newPath.startsWith("/market/orders")) {
      showOrderManagement.value = true;
    }
    // 营销互动相关路由
    if (
      newPath.startsWith("/market/marketing") ||
      newPath.startsWith("/market/personalization") ||
      newPath.startsWith("/market/cms/")
    ) {
      showMarketingInteraction.value = true;
    }
    // 售后与纠纷相关路由
    if (
      newPath.startsWith("/market/after-sales") ||
      newPath.startsWith("/market/refunds") ||
      newPath.startsWith("/market/disputes")
    ) {
      showAfterSales.value = true;
    }
    // 会员等级相关路由
    if (newPath.startsWith("/customer/membership/tiers") ||
      newPath.startsWith("/customer/membership/benefits") ||
      newPath.startsWith("/customer/membership/rules") ||
      newPath.startsWith("/customer/membership/analytics")) {
      showMembershipTiers.value = true;
    }
    // 积分与成长值相关路由
    if (newPath.startsWith("/customer/membership/points") ||
      newPath.startsWith("/customer/membership/growth-value")) {
      showMembershipPointsAndGrowth.value = true;
    }
    // 推荐/分销相关路由
    if (newPath.startsWith("/customer/affiliates")) {
      showAffiliates.value = true;
    }
    // 客户自助退换货门户相关路由
    if (newPath.startsWith("/customer/returns")) {
      showReturns.value = true;
    }
    // 属性和规格相关路由
    if (newPath.startsWith("/product/attributes") || 
        newPath.startsWith("/product/specifications") || 
        newPath.startsWith("/product/attribute-groups") || 
        newPath.startsWith("/product/category-templates")) {
      showAttributeAndSpec.value = true;
    }
  },
  {immediate: true}
);
</script>
