<template>
  <div class="store-page" :class="{ 'store-page-embedded': embeddedMode }">
    <div class="bangz_box" aria-label="快捷入口">
      <a
        v-if="complaintUrl"
        class="item item-danger"
        :href="complaintUrl"
        target="_blank"
        rel="noopener noreferrer"
      >
        <span>投诉商家</span>
      </a>

      <button
        v-if="merchantContactText"
        type="button"
        class="item item-support action-button"
        @click="copyMerchantContact"
      >
        <span>商家客服</span>
      </button>

      <a
        v-if="groupEntryUrl"
        class="item item-group"
        :href="groupEntryUrl"
        target="_blank"
        rel="noopener noreferrer"
      >
        <span>商家Q群</span>
      </a>
    </div>

    <header class="header">
      <div class="container header-inner">
        <div class="header_left">
          <button type="button" class="header_logo" @click="navigateBack">
            <img :src="headerLogoUrl" alt="LOGO" />
            <span>{{ merchantDisplayName }}</span>
          </button>

          <button
            v-if="merchantContactText"
            type="button"
            class="online-btn"
            @click="copyMerchantContact"
          >
            <svg class="icon" viewBox="0 0 1024 1024" width="16" height="16" aria-hidden="true">
              <path
                d="M877.714286 170.666667v585.142857H575.463619l-282.526476 196.851809V755.809524H146.285714V170.666667h731.428572z m-97.52381 97.523809H243.809524v390.095238h146.651428l-0.024381 107.568762L544.841143 658.285714H780.190476V268.190476z m-365.714286 121.904762a73.142857 73.142857 0 1 1 0 146.285714 73.142857 73.142857 0 0 1 0-146.285714z m195.04762 0a73.142857 73.142857 0 1 1 0 146.285714 73.142857 73.142857 0 0 1 0-146.285714z"
                fill="#ffffff"
              />
            </svg>
            <span>{{ merchantContactText }}</span>
          </button>
        </div>

        <div class="header_right">
          <a
            v-if="orderQueryUrl"
            class="header_button"
            :href="orderQueryUrl"
            target="_blank"
            rel="noopener noreferrer"
          >
            <svg class="icon" viewBox="0 0 1024 1024" width="22" height="22" aria-hidden="true">
              <path
                d="M455.253333 657.92c117.76 0 213.333333-95.573333 213.333334-213.333333s-95.573333-213.333333-213.333334-213.333334-213.333333 95.573333-213.333333 213.333334 95.573333 213.333333 213.333333 213.333333z m229.76-22.4l169.813334 169.813333c16.64 16.64 16.64 43.733333 0 60.373334-16.64 16.64-43.733333 16.64-60.373334 0l-172.8-172.8c-47.573333 32-104.746667 50.56-166.4 50.56-164.906667 0-298.666667-133.76-298.666666-298.666667s133.76-298.666667 298.666666-298.666667 298.666667 133.76 298.666667 298.666667c0 72.32-25.813333 138.88-68.906667 190.72z"
                fill="#ffffff"
              />
            </svg>
            <span>订单查询</span>
          </a>
        </div>
      </div>
    </header>

    <section class="section details__section section--first section--last">
      <div class="container merchant-card">
        <div v-if="loading && !merchantInfo" class="store-loading merchant-loading">
          <div class="loading-dot"></div>
          <span>正在同步店铺数据...</span>
        </div>

        <template v-else>
          <div class="section__title-wrap">
            <h2 class="section__title section__title--pre">商家信息</h2>
          </div>

          <hr class="mt-4" />

          <div class="merchant-meta-row">
            <div v-if="merchantInfo?.auth_status === 2" class="merchant-meta-item">
              <svg class="merchant-meta-svg" viewBox="0 0 1024 1024" aria-hidden="true">
                <path
                  d="M512 64l352 160v224c0 213.333333-146.346667 410.026667-352 448-205.653333-37.973333-352-234.666667-352-448V224L512 64z"
                  fill="#3d7dff"
                />
                <path
                  d="M734.208 364.8L473.6 625.365333 325.973333 477.738667l60.330667-60.330667L473.6 504.746667l200.277333-200.277334z"
                  fill="#ffffff"
                />
              </svg>
              <span>商家认证：<strong>已认证</strong></span>
            </div>

            <div v-if="formattedCreateDate" class="merchant-meta-item">
              <svg class="merchant-meta-svg" viewBox="0 0 1024 1024" aria-hidden="true">
                <path
                  d="M832 192H704v-64h-64v64H384v-64h-64v64H192c-35.2 0-64 28.8-64 64v512c0 35.2 28.8 64 64 64h640c35.2 0 64-28.8 64-64V256c0-35.2-28.8-64-64-64z"
                  fill="#4f7dff"
                />
                <path d="M192 352h640v416H192z" fill="#ffffff" />
                <path d="M288 448h128v96H288zM480 448h128v96H480zM672 448h64v96h-64z" fill="#9ab7ff" />
              </svg>
              <span>开店时间：{{ formattedCreateDate }}</span>
            </div>

            <div v-if="merchantContactText" class="merchant-meta-item">
              <svg class="merchant-meta-svg" viewBox="0 0 1024 1024" aria-hidden="true">
                <path
                  d="M512 64c247.424 0 448 176.608 448 394.24 0 124.416-65.792 235.392-168.32 307.776L832 960l-192.064-96.032c-40.96 8.832-83.2 13.44-127.936 13.44C264.576 877.408 64 700.8 64 458.24S264.576 64 512 64z"
                  fill="#30c56d"
                />
                <path
                  d="M359.744 426.24c25.152-51.456 59.392-77.504 102.656-77.504 23.04 0 46.08 7.68 69.632 23.04 16.896 11.264 25.088 24.576 25.088 40.448 0 11.264-4.608 24.064-14.336 38.4l-23.04 33.28c17.408 24.576 49.152 53.248 95.744 86.016l31.232-21.504c13.824-9.728 27.648-14.336 41.472-14.336 15.36 0 29.184 6.656 40.448 20.48 18.944 22.528 28.16 45.568 28.16 68.608 0 37.888-23.552 67.072-70.144 87.552-16.896 7.68-34.304 11.264-52.224 11.264-38.4 0-83.456-16.384-135.168-49.664-61.44-39.936-105.984-87.552-133.632-142.336-13.824-27.648-20.992-53.76-20.992-78.272 0-9.216 1.024-17.92 3.584-26.624z"
                  fill="#ffffff"
                />
              </svg>
              <span>{{ merchantContactText }}</span>
            </div>
          </div>
        </template>
      </div>

      <div class="container main-card">
        <div v-if="loading" class="store-loading">
          <div class="loading-dot"></div>
          <span>正在同步店铺数据...</span>
        </div>

        <template v-else>
          <div class="row category">
            <div class="col-12">
              <div class="section__title-wrap">
                <h2 class="section__title section__title2">商品搜索</h2>
              </div>
            </div>

            <div class="col-12">
              <div class="section__title-wrap search">
                <div class="input">
                  <input
                    v-model="searchInput"
                    type="text"
                    placeholder="输入关键词搜索商品"
                    @keyup.enter="runSearch"
                  />
                </div>

                <button type="button" class="search_button" @click="runSearch">
                  <svg class="icon" viewBox="0 0 1024 1024" width="22" height="22" aria-hidden="true">
                    <path
                      d="M455.253333 657.92c117.76 0 213.333333-95.573333 213.333334-213.333333s-95.573333-213.333333-213.333334-213.333334-213.333333 95.573333-213.333333 213.333334 95.573333 213.333333 213.333333 213.333333z m229.76-22.4l169.813334 169.813333c16.64 16.64 16.64 43.733333 0 60.373334-16.64 16.64-43.733333 16.64-60.373334 0l-172.8-172.8c-47.573333 32-104.746667 50.56-166.4 50.56-164.906667 0-298.666667-133.76-298.666666-298.666667s133.76-298.666667 298.666666-298.666667 298.666667 133.76 298.666667 298.666667c0 72.32-25.813333 138.88-68.906667 190.72z"
                      fill="#ffffff"
                    />
                  </svg>
                  <span>商品查询</span>
                </button>
              </div>
            </div>

            <div class="col-12">
              <div class="section__title-wrap">
                <h2 class="section__title">商品分类</h2>
              </div>
            </div>

            <div class="col-12 col-lg-12">
              <div class="category_list">
                <button
                  v-for="category in categories"
                  :key="category.id"
                  type="button"
                  class="category_box"
                  :class="{ active: !activeKeywords && category.id === selectedCategoryId }"
                  @click="selectCategory(category)"
                >
                  <div class="card">
                    <div class="card__title pb-0">
                      <h3>{{ category.name }}</h3>
                    </div>
                    <div class="card__content">
                      <span class="card__s_cateremark">共{{ category.goods_count }}种商品</span>
                    </div>
                    <span
                      v-if="!activeKeywords && category.id === selectedCategoryId"
                      class="category-lite_img"
                      aria-hidden="true"
                    ></span>
                  </div>
                </button>
              </div>
            </div>
          </div>

          <hr class="mt-4" />

          <div class="row goods">
            <div class="col-12">
              <div class="section__title-wrap">
                <h2 class="section__title section__title--pre" id="goods_title">
                  {{ activeKeywords ? '搜索结果' : '选择商品' }}
                </h2>
              </div>
            </div>

            <div class="col-12 col-lg-12">
              <div v-if="goodsLoading" class="panel-empty">正在加载商品...</div>
              <div v-else-if="goodsList.length === 0" class="panel-empty">暂无商品</div>
              <div v-else class="goods_list">
                <button
                  v-for="goods in goodsList"
                  :key="goods.goods_key"
                  type="button"
                  class="goods_box"
                  :class="{ active: goods.goods_key === selectedProductKey }"
                  @click="selectProduct(goods)"
                >
                  <div class="card">
                    <div class="card__title">
                      <h3>{{ goods.name }}</h3>
                    </div>
                    <div class="card__detail">
                      <span class="card__detail_price">￥{{ formatAmount(goods.price) }}</span>
                      <span class="card__detail_stock">{{ resolveGoodsSubline(goods) }}</span>
                    </div>
                    <span
                      v-if="goods.goods_key === selectedProductKey"
                      class="goods-lite_img"
                      aria-hidden="true"
                    ></span>
                  </div>
                </button>
              </div>
            </div>
          </div>

          <div v-if="saleMessages.length" class="mt-4">
            <div class="d-flex mt-0 align-items-center title-row">
              <svg class="icon" viewBox="0 0 1024 1024" width="18" height="18" aria-hidden="true">
                <path
                  d="M988.908308 614.006154c-58.525538-29.538462-76.091077-53.208615-76.091077-100.509539 0-44.386462 17.565538-68.017231 76.091077-100.548923 32.177231-20.716308 35.091692-38.439385 35.091692-68.01723v-168.566154C1024 123.116308 980.125538 78.769231 927.468308 78.769231H96.492308C43.874462 78.769231 0 123.116308 0 176.364308v168.566154c0 26.584615 2.914462 50.254769 35.091692 68.01723 23.433846 11.815385 76.091077 41.353846 76.091077 100.548923 0 65.024-38.045538 85.740308-73.137231 97.555693H35.052308C2.914462 631.768615 0 664.300308 0 679.069538v168.566154C0 900.883692 43.874462 945.230769 96.531692 945.230769H927.507692C980.125538 945.230769 1024 900.883692 1024 847.635692v-168.566154c0-32.531692-11.697231-44.347077-35.091692-65.063384z"
                  fill="#7FBCFF"
                />
                <path
                  d="M670.444308 530.116923c17.723077 0 32.571077 14.572308 32.571077 32.019692a32.571077 32.571077 0 0 1-32.571077 31.980308h-124.376616v122.171077a32.571077 32.571077 0 0 1-65.142154 0v-122.171077H356.548923a32.571077 32.571077 0 0 1-32.571077-31.980308c0-17.447385 14.808615-32.019692 32.571077-32.019692h124.376615v-75.618461H347.648A32.571077 32.571077 0 0 1 315.076923 422.478769c0-17.447385 14.808615-31.980308 32.571077-31.980307h97.713231L341.740308 288.689231a31.232 31.232 0 0 1 0-43.638154 32.610462 32.610462 0 0 1 44.425846 0l127.330461 125.085538 127.330462-125.085538a32.610462 32.610462 0 0 1 44.386461 0c11.854769 11.618462 11.854769 31.980308 0 43.638154l-103.620923 101.809231h94.759385c17.762462 0 32.571077 14.572308 32.571077 31.980307a32.571077 32.571077 0 0 1-32.571077 32.019693h-133.277538v75.618461h127.369846z"
                  fill="#007AFF"
                />
              </svg>
              <h2 class="section__title_2 mb-0 ml-2">商品优惠</h2>
            </div>
            <div class="text_box sale_message">
              <span v-for="message in saleMessages" :key="message">{{ message }}</span>
            </div>
          </div>

          <div class="mt-4">
            <div class="d-flex mt-0 align-items-center title-row">
              <svg class="icon" viewBox="0 0 1024 1024" width="18" height="18" aria-hidden="true">
                <path
                  d="M392.7 958.9c22.5 0 40.7-18.2 40.7-40.7V630.9c0-22.5-18.2-40.7-40.7-40.7H105.4c-22.5 0-40.7 18.2-40.7 40.7v287.4c0 22.5 18.2 40.7 40.7 40.7h287.3zM860 958.9c22.5 0 40.7-18.2 40.7-40.7V630.9c0-22.5-18.2-40.7-40.7-40.7H572.7c-22.5 0-40.7 18.2-40.7 40.7v287.4c0 22.5 18.2 40.7 40.7 40.7H860zM392.7 492c22.5 0 40.7-18.2 40.7-40.7V164c0-22.5-18.2-40.7-40.7-40.7H105.4c-22.5 0-40.7 18.2-40.7 40.7v287.4c0 22.5 18.2 40.7 40.7 40.7h287.3z"
                  fill="#1E94EE"
                />
                <path
                  d="M948.3 336.4c15.9-15.9 15.9-41.6 0-57.5L745.1 75.7c-15.9-15.9-41.6-15.9-57.5 0L484.4 278.9c-15.9 15.9-15.9 41.6 0 57.5l203.2 203.2c15.9 15.9 41.6 15.9 57.5 0l203.2-203.2z"
                  fill="#B4DCF5"
                />
              </svg>
              <h2 class="section__title_2 mb-0 ml-2">商品描述</h2>
            </div>
            <div class="text_box mt-3" id="remark">
              <p>{{ selectedProductDescription }}</p>
            </div>
          </div>

          <hr class="mt-4" />

          <div class="row mt-4">
            <div class="col-12" id="order_box">
              <div v-if="subscriptionGoodsSelected" class="ure_info_box subscription-entry-box">
                <div class="ure_info_hide">
                  <span>订阅开通说明</span>
                </div>

                <div class="subscription-entry-content">
                  <p class="subscription-entry-title">{{ selectedProduct?.name }}</p>
                  <p>该商品属于站内订阅套餐，点击下方按钮后会跳转到订阅支付页继续完成开通。</p>
                  <p v-if="!internalPaymentEnabled" class="subscription-entry-hint">
                    当前站内支付尚未开启，套餐先展示在本页，开启后即可直接购买。
                  </p>
                </div>
              </div>

              <div v-else class="ure_info_box">
                <div class="ure_info_hide">
                  <span>填写信息/购买方式</span>
                </div>

                <div class="ure_info">
                  <div class="ure_info_input_box">
                    <div class="ure_info_input_box_hide">
                      <p>*</p>
                      <p>联系方式</p>
                    </div>
                    <div class="input">
                      <input
                        v-model="contact"
                        type="text"
                        placeholder="[必填]请填写 Bridgemind 注册邮箱"
                      />
                    </div>
                    <div class="msg">联系方式特别重要，可用来查询订单与自动识别充值账号</div>
                  </div>

                  <div class="ure_info_input_box d-flex purchase-toggle-row">
                    <div class="ure_info_input_box_hide btn-type">
                      <div :class="{ on: smsReminderEnabled }" @click="smsReminderEnabled = !smsReminderEnabled">
                        <label>短信提醒</label>
                      </div>
                    </div>

                    <div class="ure_info_input_box_hide btn-type">
                      <div :class="{ on: emailReminderEnabled }" @click="emailReminderEnabled = !emailReminderEnabled">
                        <label>邮箱提醒</label>
                      </div>
                    </div>

                    <div class="ure_info_input_box_hide btn-type">
                      <div
                        :class="{ on: couponEnabled, disabled: !canUseCoupon }"
                        @click="toggleCoupon"
                      >
                        <label>使用优惠券</label>
                      </div>
                    </div>
                  </div>

                  <div v-if="couponEnabled && canUseCoupon" class="ure_info_input_box">
                    <div class="input">
                      <input v-model="couponCode" type="text" placeholder="请填写你的优惠券" />
                    </div>
                  </div>
                </div>

                <div class="pay_type">
                  <div class="pay_type_hide">选择支付方式</div>
                  <div class="pay_type_box">
                    <button
                      v-for="(channel, index) in channels"
                      :key="channel.id"
                      type="button"
                      class="pay_type_leng"
                      :class="{ pay_type_leng_xz: channel.id === selectedChannelId }"
                      @click="selectedChannelId = channel.id"
                    >
                      <img
                        v-if="channel.paytype?.icon"
                        :src="channel.paytype.icon"
                        :alt="channel.show_name"
                      />
                      <span>{{ formatChannelLabel(channel, index) }}</span>
                      <span
                        v-if="channel.id === selectedChannelId"
                        class="xiadui"
                        aria-hidden="true"
                      ></span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>
    </section>

    <footer>
      <div class="container">
        <p>Copyright © 2024-2026 {{ siteName }} All rights reserved. 版权所有</p>
      </div>
    </footer>

    <section class="section_buttom">
      <div class="container bottom-inner">
        <div class="goods_name">
          <svg class="icon" viewBox="0 0 1024 1024" width="28" height="28" aria-hidden="true">
            <path
              d="M847.872 734.208L784.384 368.64H239.616l-63.488 365.568c-9.216 51.2 30.72 105.472 81.92 105.472h506.88c52.224 0 92.16-54.272 82.944-105.472zM512 648.192c-71.68 0-130.048-58.368-130.048-130.048 0-6.144 4.096-10.24 10.24-10.24s10.24 4.096 10.24 10.24c0 60.416 49.152 109.568 109.568 109.568s109.568-49.152 109.568-109.568c0-6.144 4.096-10.24 10.24-10.24s10.24 4.096 10.24 10.24c0 71.68-58.368 130.048-130.048 130.048z"
              fill="#2F8AF5"
            />
            <path
              d="M791.552 358.4l-19.456-96.256c-7.168-39.936-46.08-67.584-87.04-67.584H340.992c-40.96 0-74.752 28.672-81.92 68.608L243.712 358.4h547.84z"
              fill="#83C1FF"
            />
          </svg>
          <span>{{ selectedProduct?.name || '请选择商品' }}</span>
        </div>

        <div class="shuliang_box">
          <template v-if="subscriptionGoodsSelected">
            <div class="subscription-bottom-summary">
              <span>{{ resolveGoodsSubline(selectedProduct!) }}</span>
              <span v-if="selectedProduct?.subscription_plan?.group_name">
                {{ selectedProduct.subscription_plan.group_name }}
              </span>
            </div>
          </template>

          <template v-else>
            <button type="button" class="btn" @click="decreaseQuantity">
              <span class="minus-bar"></span>
            </button>

            <div class="input">
              <input
                v-model.number="quantity"
                type="number"
                min="1"
                inputmode="numeric"
                @blur="normalizeQuantity"
                @keyup.enter="normalizeQuantity"
              />
            </div>

            <button type="button" class="btn" @click="increaseQuantity">
              <span></span>
              <span></span>
            </button>
          </template>

          <div class="jiage">
            <span>{{ subscriptionGoodsSelected ? '套餐金额 :' : '支付金额 :' }}</span>
            <span>￥</span>
            <s v-if="showOriginalAmount">{{ formatAmount(originalAmount) }}</s>
            <span>{{ formatAmount(totalAmount) }}</span>
            <span v-if="paymentFeeText" class="payment-fee-text">{{ paymentFeeText }}</span>
          </div>
        </div>

        <div class="queding_btn">
          <button
            type="button"
            name="check_pay"
            class="check_pay"
            :disabled="ordering || !selectedProduct"
            @click="submitOrder"
          >
            {{ submitButtonText }}
          </button>
        </div>
      </div>
    </section>

    <div v-if="qrCodeDataUrl" class="ewm">
      <img :src="qrCodeDataUrl" alt="支付二维码" />
      手机扫码支付
    </div>

    <div v-if="agreementVisible" class="agreement-overlay">
      <div class="agreement-modal">
        <div class="agreement-header">
          <h3>购卡协议</h3>
          <button type="button" class="agreement-close" @click="closeAgreement">×</button>
        </div>

        <div class="agreement-body">
          <p><strong>温馨提示：</strong>本站不提供任何担保，私下交易被骗一律与本站无关。</p>
          <p class="agreement-emphasis">
            提示：有问题左上角联系客服，商品存在争议请投诉商家
          </p>
          <ol>
            <li>平台为隔天中午12点结算，商品有问题请于当天24点前及时联系商家。</li>
            <li>平台仅提供自动发卡、寄售服务，非商品本身问题不受理售后争议。</li>
            <li>订单查询期为1个月，购买后请自行保存关键信息。</li>
          </ol>
        </div>

        <div class="agreement-footer">
          <button type="button" @click="acceptAgreement">我同意协议</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import QRCode from 'qrcode'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import type { SubscriptionPlan as PaymentSubscriptionPlan } from '@/types/payment'

interface ApiResponse<T> {
  code: number
  msg: string
  data: T
}

interface ShopCategory {
  id: number
  name: string
  image?: string
  goods_count: number
  synthetic_type?: 'subscription'
}

interface ShopInfo {
  create_time: number
  nickname: string
  avatar?: string
  description: string
  token: string
  auth_status: number
  contact_qq?: string
  contact_mobile?: string
  contact_wechat?: string
}

interface ShopUserInfo {
  link?: string
  nickname?: string
  avatar?: string
  token?: string
}

interface ShopGoodsCategory {
  id: number
  name: string
}

interface ShopGoodsExtend {
  stock_count?: number
  show_stock_type?: number
  send_order?: number
  limit_count?: number
  query_password_status?: number
}

interface ShopGoodsDiscount {
  available: number
  rebate: number
}

interface ShopGoods {
  link: string
  goods_key: string
  name: string
  price: number
  market_price: number
  description: string
  image?: string
  contact_format?: string
  coupon_status?: number
  category?: ShopGoodsCategory
  user?: ShopUserInfo
  discount?: ShopGoodsDiscount | null
  multipleoffers?: { available: number; discount_type: number } | null
  fullgift?: { available: number } | null
  extend?: ShopGoodsExtend | null
  synthetic_type?: 'subscription'
  subscription_plan_id?: number
  subscription_group_id?: number
  subscription_plan?: SubscriptionPlan
}

interface GoodsListPayload {
  total: number
  list: ShopGoods[]
}

interface ShopChannel {
  id: number
  show_name: string
  rate?: number
  paytype?: {
    icon?: string
  } | null
}

interface PriceInfo {
  original_amount: number
  total_amount: number
  fee: number
  fee_payer: number
  sales_style?: Record<string, string> | string[]
  coupon_available?: number
  coupon_price?: number
}

interface CreateOrderResponse {
  payurl?: string
  trade_no?: string
  total_amount: number
}

type SubscriptionPlan = PaymentSubscriptionPlan & {
  product_name?: string
}

const DEFAULT_SHOP_URL = 'https://pay.ldxp.cn/shop/HQP8RZ4F'
const AGREEMENT_STORAGE_KEY = 'bridgemind.recharge.agreement.v3'
const DEFAULT_HEADER_LOGO = 'https://happp.cn/static/app/theme/default/img/shop_img.png'
const SUBSCRIPTION_CATEGORY_ID = -20260423

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const loading = ref(true)
const goodsLoading = ref(false)
const ordering = ref(false)
const agreementVisible = ref(false)

const merchantInfo = ref<ShopInfo | null>(null)
const categories = ref<ShopCategory[]>([])
const goodsList = ref<ShopGoods[]>([])
const channels = ref<ShopChannel[]>([])
const priceInfo = ref<PriceInfo | null>(null)
const subscriptionPlans = ref<SubscriptionPlan[]>([])

const selectedCategoryId = ref<number | null>(null)
const selectedProductKey = ref('')
const selectedChannelId = ref<number | null>(null)

const quantity = ref(1)
const contact = ref('')
const couponEnabled = ref(false)
const couponCode = ref('')
const searchInput = ref('')
const activeKeywords = ref('')
const smsReminderEnabled = ref(false)
const emailReminderEnabled = ref(false)
const qrCodeDataUrl = ref('')

let latestPriceRequestId = 0

const embeddedMode = computed(() => {
  if (route.query.ui_mode === 'embedded') return true
  if (typeof window === 'undefined') return false
  return window.self !== window.top
})

const siteName = computed(() => appStore.siteName || appStore.cachedPublicSettings?.site_name || 'Bridgemind')
const internalPaymentEnabled = computed(() => Boolean(appStore.cachedPublicSettings?.payment_enabled))
const hasSubscriptionGoods = computed(() => subscriptionPlans.value.length > 0)

const sourceUrl = computed(() => {
  const raw = typeof route.query.src_url === 'string' ? route.query.src_url : ''
  return sanitizeURL(raw, '')
})

const shopUrl = computed(() => {
  const raw = typeof route.query.shop_url === 'string' ? route.query.shop_url : DEFAULT_SHOP_URL
  return sanitizeURL(raw, DEFAULT_SHOP_URL)
})

const shopOrigin = computed(() => {
  try {
    return new URL(shopUrl.value).origin
  } catch {
    return 'https://pay.ldxp.cn'
  }
})

const shopToken = computed(() => {
  if (typeof route.query.shop_token === 'string' && route.query.shop_token.trim()) {
    return route.query.shop_token.trim()
  }
  return extractShopToken(shopUrl.value)
})

const merchantDisplayName = computed(() => merchantInfo.value?.nickname?.trim() || siteName.value)
const headerLogoUrl = computed(() => DEFAULT_HEADER_LOGO)
const orderQueryUrl = computed(() => buildShopUrl('/orderquery'))
const complaintUrl = computed(() => buildShopUrl('/complaint'))
const groupEntryUrl = computed(() => sourceUrl.value || shopUrl.value)

const merchantContactValue = computed(() => {
  const info = merchantInfo.value
  return (
    info?.contact_wechat?.trim() ||
    info?.contact_qq?.trim() ||
    appStore.contactInfo?.trim() ||
    ''
  )
})

const merchantContactLabel = computed(() => {
  const info = merchantInfo.value
  if (info?.contact_wechat?.trim()) return '商家微信'
  if (info?.contact_qq?.trim()) return '商家QQ'
  return '站点联系'
})

const merchantContactText = computed(() =>
  merchantContactValue.value ? `${merchantContactLabel.value}：${merchantContactValue.value}` : ''
)

const formattedCreateDate = computed(() =>
  merchantInfo.value?.create_time ? formatUnixDate(merchantInfo.value.create_time) : ''
)

const selectedProduct = computed(
  () => goodsList.value.find((goods) => goods.goods_key === selectedProductKey.value) || null
)

const subscriptionGoodsSelected = computed(() => isSubscriptionGoods(selectedProduct.value))

const selectedProductDescription = computed(() => {
  if (isSubscriptionGoods(selectedProduct.value)) {
    return buildSubscriptionDescription(selectedProduct.value.subscription_plan)
  }

  const description = stripHtml(selectedProduct.value?.description || '')
  if (description) return description
  return '请选择需要充值的商品，填写 Bridgemind 注册邮箱后即可发起支付。'
})

const minQuantity = computed(() => Math.max(1, selectedProduct.value?.extend?.limit_count || 1))
const canUseCoupon = computed(() => (selectedProduct.value?.coupon_status ?? 0) === 1)

const totalAmount = computed(() => {
  if (priceInfo.value) return priceInfo.value.total_amount
  if (!selectedProduct.value) return 0
  return selectedProduct.value.price * quantity.value
})

const originalAmount = computed(() => {
  if (priceInfo.value) return priceInfo.value.original_amount
  if (isSubscriptionGoods(selectedProduct.value)) {
    return Number(selectedProduct.value.market_price || selectedProduct.value.price || 0)
  }
  return totalAmount.value
})

const showOriginalAmount = computed(() => originalAmount.value > totalAmount.value)

const paymentFeeText = computed(() => {
  const fee = Number(priceInfo.value?.fee ?? 0)
  if (fee <= 0) return ''
  return `含手续费：${formatAmount(fee)}元`
})

const submitButtonText = computed(() => {
  if (ordering.value) {
    return subscriptionGoodsSelected.value ? '正在跳转订阅页' : '正在创建订单'
  }

  if (subscriptionGoodsSelected.value) {
    return internalPaymentEnabled.value ? '立即开通订阅' : '订阅暂未开放'
  }

  return '确认支付'
})

const saleMessages = computed(() => {
  const salesStyle = priceInfo.value?.sales_style
  if (!salesStyle) return []
  if (Array.isArray(salesStyle)) {
    return salesStyle.map((item) => String(item))
  }
  return Object.entries(salesStyle).map(([key, value]) => `${key} ${value}`.trim())
})

function sanitizeURL(raw: string, fallback: string): string {
  if (!raw) return fallback
  try {
    const url = new URL(raw, window.location.origin)
    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return fallback
    }
    return url.toString()
  } catch {
    return fallback
  }
}

function extractShopToken(url: string): string {
  try {
    const parsed = new URL(url)
    const parts = parsed.pathname.split('/').filter(Boolean)
    const shopIndex = parts.findIndex((part) => part === 'shop')
    if (shopIndex >= 0 && parts[shopIndex + 1]) {
      return parts[shopIndex + 1]
    }
  } catch {
    return ''
  }
  return ''
}

function buildShopUrl(path: string): string {
  try {
    return new URL(path, shopOrigin.value).toString()
  } catch {
    return shopUrl.value
  }
}

function stripHtml(value: string): string {
  return value
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function formatUnixDate(unixSeconds: number): string {
  const date = new Date(unixSeconds * 1000)
  if (Number.isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${year}-${month}-${day}`
}

function formatAmount(value: number): string {
  return Number(value || 0).toFixed(2)
}

function isSubscriptionGoods(goods: ShopGoods | null | undefined): goods is ShopGoods & {
  synthetic_type: 'subscription'
  subscription_plan: SubscriptionPlan
  subscription_plan_id: number
} {
  return Boolean(goods?.synthetic_type === 'subscription' && goods.subscription_plan)
}

function formatSubscriptionValidity(plan: SubscriptionPlan): string {
  const unit = plan.validity_unit || 'day'
  if (unit === 'month') return '包月订阅'
  if (unit === 'year') return '包年订阅'
  if (plan.validity_days > 0) return `${plan.validity_days}天订阅`
  return '订阅套餐'
}

function buildSubscriptionDescription(plan: SubscriptionPlan): string {
  const parts = [
    plan.description?.trim(),
    `有效期：${formatSubscriptionValidity(plan)}`,
    plan.group_name ? `所属分组：${plan.group_name}` : '',
    plan.rate_multiplier ? `倍率：×${Number(plan.rate_multiplier.toPrecision(10))}` : '',
    plan.daily_limit_usd != null ? `日额度：$${plan.daily_limit_usd}` : '',
    plan.weekly_limit_usd != null ? `周额度：$${plan.weekly_limit_usd}` : '',
    plan.monthly_limit_usd != null ? `月额度：$${plan.monthly_limit_usd}` : '',
    ...(plan.features || []).filter(Boolean),
  ].filter(Boolean)

  return parts.join('\n')
}

function buildSubscriptionGoods(plan: SubscriptionPlan): ShopGoods {
  const displayName = plan.product_name?.trim() || plan.name
  return {
    link: '',
    goods_key: `subscription-plan-${plan.id}`,
    name: displayName,
    price: Number(plan.price || 0),
    market_price: Number(plan.original_price || plan.price || 0),
    description: buildSubscriptionDescription(plan),
    contact_format: 'email',
    coupon_status: 0,
    category: {
      id: SUBSCRIPTION_CATEGORY_ID,
      name: '订阅',
    },
    extend: {
      stock_count: 9999,
      send_order: 0,
      limit_count: 1,
    },
    synthetic_type: 'subscription',
    subscription_plan_id: plan.id,
    subscription_group_id: plan.group_id,
    subscription_plan: plan,
  }
}

function buildSubscriptionGoodsList(): ShopGoods[] {
  return subscriptionPlans.value.map((plan) => buildSubscriptionGoods(plan))
}

function filterSubscriptionGoodsByKeyword(keyword: string): ShopGoods[] {
  const normalizedKeyword = keyword.trim().toLowerCase()
  if (!normalizedKeyword) return buildSubscriptionGoodsList()

  return buildSubscriptionGoodsList().filter((goods) => {
    const haystacks = [
      goods.name,
      goods.description,
      goods.subscription_plan?.name,
      goods.subscription_plan?.product_name,
      goods.subscription_plan?.group_name,
    ]

    return haystacks.some((value) => value?.toLowerCase().includes(normalizedKeyword))
  })
}

function formatChannelLabel(channel: ShopChannel, index: number): string {
  const base = channel.show_name?.trim() || '支付方式'
  if (index === 0 && !base.startsWith('[推荐]')) {
    return `[推荐]${base}`
  }
  return base
}

function resolveGoodsSubline(goods: ShopGoods): string {
  if (isSubscriptionGoods(goods)) {
    return formatSubscriptionValidity(goods.subscription_plan)
  }

  if ((goods.contact_format || '').toLowerCase() === 'email') {
    return '填写注册邮箱'
  }

  const stockCount = goods.extend?.stock_count ?? 0
  if (stockCount > 0) {
    return `库存${stockCount}份`
  }

  return '自动发货'
}

function navigateInternal(path: string) {
  if (typeof window !== 'undefined' && embeddedMode.value && window.top && window.top !== window) {
    window.top.location.href = new URL(path, window.location.origin).toString()
    return
  }
  router.push(path)
}

function openExternal(url: string, preferNewTab: boolean = false) {
  if (!url || typeof window === 'undefined') return

  if (preferNewTab && !embeddedMode.value) {
    const opened = window.open(url, '_blank', 'noopener,noreferrer')
    if (opened) {
      opened.opener = null
      return
    }
  }

  if (embeddedMode.value && window.top && window.top !== window) {
    window.top.location.href = url
    return
  }

  window.location.href = url
}

function navigateBack() {
  if (sourceUrl.value) {
    openExternal(sourceUrl.value, false)
    return
  }
  navigateInternal('/dashboard')
}

async function refreshShareQrCode() {
  if (typeof window === 'undefined') return

  try {
    qrCodeDataUrl.value = await QRCode.toDataURL(window.location.href, {
      width: 150,
      margin: 1,
    })
  } catch {
    qrCodeDataUrl.value = ''
  }
}

async function postShopApi<T>(path: string, payload: Record<string, unknown>): Promise<T> {
  const response = await fetch(buildShopUrl(path), {
    method: 'POST',
    mode: 'cors',
    credentials: 'omit',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    throw new Error(`请求失败（${response.status}）`)
  }

  const result = (await response.json()) as ApiResponse<T>
  if (result.code !== 1) {
    throw new Error(result.msg || '请求失败')
  }

  return result.data
}

async function loadMerchantInfo() {
  merchantInfo.value = await postShopApi<ShopInfo>('/shopApi/Shop/info', {
    token: shopToken.value,
  })
}

async function loadChannels() {
  channels.value = await postShopApi<ShopChannel[]>('/shopApi/Shop/getUserChannel', {
    token: shopToken.value,
  })
  selectedChannelId.value = channels.value[0]?.id ?? null
}

async function loadSubscriptionPlans() {
  try {
    const response = await paymentAPI.getCheckoutInfo()
    const plans = Array.isArray(response.data?.plans) ? response.data.plans : []
    subscriptionPlans.value = [...plans].sort((a, b) => {
      const left = typeof a.sort_order === 'number' ? a.sort_order : 0
      const right = typeof b.sort_order === 'number' ? b.sort_order : 0
      return left - right
    })
  } catch {
    subscriptionPlans.value = []
  }
}

async function loadCategories() {
  const remoteCategories = await postShopApi<ShopCategory[]>('/shopApi/Shop/categoryList', {
    token: shopToken.value,
    goods_type: 'card',
  })

  categories.value = [...remoteCategories]
  if (hasSubscriptionGoods.value) {
    categories.value.push({
      id: SUBSCRIPTION_CATEGORY_ID,
      name: '订阅',
      goods_count: subscriptionPlans.value.length,
      synthetic_type: 'subscription',
    })
  }

  if (!categories.value.length) {
    selectedCategoryId.value = null
    goodsList.value = []
    selectedProductKey.value = ''
    return
  }

  if (
    !selectedCategoryId.value ||
    !categories.value.some((category) => category.id === selectedCategoryId.value)
  ) {
    selectedCategoryId.value = categories.value[0].id
  }

  await loadGoods()
}

async function loadGoods() {
  goodsLoading.value = true

  try {
    if (!activeKeywords.value && selectedCategoryId.value === SUBSCRIPTION_CATEGORY_ID) {
      goodsList.value = buildSubscriptionGoodsList()
      if (!goodsList.value.some((goods) => goods.goods_key === selectedProductKey.value)) {
        selectedProductKey.value = goodsList.value[0]?.goods_key ?? ''
      }
      return
    }

    const payload: Record<string, unknown> = {
      token: shopToken.value,
      goods_type: 'card',
      current: 1,
      pageSize: 999,
      keywords: activeKeywords.value,
    }

    if (!activeKeywords.value && selectedCategoryId.value !== null) {
      payload.category_id = selectedCategoryId.value
    }

    const goodsPayload = await postShopApi<GoodsListPayload>('/shopApi/Shop/goodsList', payload)
    const remoteGoods = goodsPayload.list || []
    const subscriptionGoods = activeKeywords.value ? filterSubscriptionGoodsByKeyword(activeKeywords.value) : []
    goodsList.value = [...subscriptionGoods, ...remoteGoods]

    if (!goodsList.value.some((goods) => goods.goods_key === selectedProductKey.value)) {
      selectedProductKey.value = goodsList.value[0]?.goods_key ?? ''
    }
  } finally {
    goodsLoading.value = false
  }
}

async function loadPriceInfo() {
  const product = selectedProduct.value
  if (!product) {
    priceInfo.value = null
    return
  }

  if (isSubscriptionGoods(product)) {
    priceInfo.value = null
    return
  }

  if (!selectedChannelId.value) {
    priceInfo.value = null
    return
  }

  const requestId = ++latestPriceRequestId

  try {
    const data = await postShopApi<PriceInfo>('/shopApi/Shop/getGoodsPrice', {
      goods_key: product.goods_key,
      quantity: quantity.value,
      coupon_code: couponEnabled.value ? couponCode.value.trim() : '',
      channel_id: selectedChannelId.value,
    })

    if (requestId !== latestPriceRequestId) return
    priceInfo.value = data
  } catch (error) {
    if (requestId !== latestPriceRequestId) return
    priceInfo.value = null
    if (couponEnabled.value && couponCode.value.trim()) {
      appStore.showError((error as Error).message || '优惠券不可用')
    }
  }
}

function normalizeQuantity() {
  const parsed = Number.parseInt(String(quantity.value), 10)
  quantity.value = Number.isFinite(parsed) ? Math.max(minQuantity.value, parsed) : minQuantity.value
}

function increaseQuantity() {
  quantity.value += 1
}

function decreaseQuantity() {
  quantity.value = Math.max(minQuantity.value, quantity.value - 1)
}

function selectCategory(category: ShopCategory) {
  activeKeywords.value = ''
  searchInput.value = ''
  selectedCategoryId.value = category.id
  void loadGoods().catch((error: unknown) => {
    appStore.showError((error as Error).message || '加载商品失败')
  })
}

function selectProduct(goods: ShopGoods) {
  selectedProductKey.value = goods.goods_key
}

function runSearch() {
  activeKeywords.value = searchInput.value.trim()
  void loadGoods().catch((error: unknown) => {
    appStore.showError((error as Error).message || '加载商品失败')
  })
}

function toggleCoupon() {
  if (!canUseCoupon.value) {
    appStore.showInfo('当前商品暂不支持优惠券')
    return
  }
  couponEnabled.value = !couponEnabled.value
}

function closeAgreement() {
  agreementVisible.value = false
}

function acceptAgreement() {
  agreementVisible.value = false
  if (typeof window !== 'undefined') {
    window.sessionStorage.setItem(AGREEMENT_STORAGE_KEY, '1')
  }
}

async function copyMerchantContact() {
  if (!merchantContactValue.value) return

  try {
    await navigator.clipboard.writeText(merchantContactValue.value)
    appStore.showSuccess('联系方式已复制')
  } catch {
    appStore.showInfo(merchantContactText.value)
  }
}

async function submitOrder() {
  if (!selectedProduct.value) {
    appStore.showError('请先选择商品')
    return
  }

  if (isSubscriptionGoods(selectedProduct.value)) {
    if (!internalPaymentEnabled.value) {
      appStore.showInfo('订阅商品已展示，当前站内支付尚未开启，开启后即可直接购买')
      return
    }

    const planId = selectedProduct.value.subscription_plan_id
    if (!planId) {
      appStore.showError('未找到订阅套餐信息')
      return
    }

    ordering.value = true
    try {
      navigateInternal(`/purchase?tab=subscription&plan=${planId}`)
    } finally {
      ordering.value = false
    }
    return
  }

  const trimmedContact = contact.value.trim()
  if (!trimmedContact) {
    appStore.showError('请填写 Bridgemind 注册邮箱')
    return
  }

  if (
    (selectedProduct.value.contact_format || '').toLowerCase() === 'email' &&
    !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedContact)
  ) {
    appStore.showError('请输入有效的邮箱地址')
    return
  }

  if (!selectedChannelId.value) {
    appStore.showError('请选择支付方式')
    return
  }

  ordering.value = true

  try {
    const payload: Record<string, unknown> = {
      goods_key: selectedProduct.value.goods_key,
      quantity: quantity.value,
      coupon_code: couponEnabled.value ? couponCode.value.trim() : '',
      channel_id: selectedChannelId.value,
      contact: trimmedContact,
      is_rev_sms: smsReminderEnabled.value ? 1 : 0,
      isemail: emailReminderEnabled.value ? 1 : 0,
    }

    if (emailReminderEnabled.value) {
      payload.email = trimmedContact
    }

    const data = await postShopApi<CreateOrderResponse>('/shopApi/Pay/order', payload)

    if (data.total_amount === 0 && data.trade_no) {
      appStore.showSuccess('订单已创建')
      openExternal(buildShopUrl(`/order/result/${data.trade_no}`), false)
      return
    }

    if (!data.payurl) {
      throw new Error('支付链接生成失败')
    }

    appStore.showSuccess('订单已创建，正在打开支付页面')
    openExternal(data.payurl, true)
  } catch (error) {
    appStore.showError((error as Error).message || '创建订单失败')
  } finally {
    ordering.value = false
  }
}

async function initializeStore() {
  if (!shopToken.value) {
    appStore.showError('未找到店铺配置')
    loading.value = false
    return
  }

  loading.value = true

  try {
    await Promise.all([loadMerchantInfo(), loadChannels(), loadSubscriptionPlans()])
    await loadCategories()
    await loadPriceInfo()
  } catch (error) {
    appStore.showError((error as Error).message || '同步店铺失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => authStore.user?.email,
  (value) => {
    if (!contact.value && value) {
      contact.value = value
    }
  },
  { immediate: true },
)

watch(selectedProduct, (product) => {
  if (!product) {
    priceInfo.value = null
    return
  }

  if (isSubscriptionGoods(product)) {
    quantity.value = 1
  }

  if (quantity.value < minQuantity.value) {
    quantity.value = minQuantity.value
  }

  if (!canUseCoupon.value) {
    couponEnabled.value = false
    couponCode.value = ''
  }
})

watch(couponEnabled, (enabled) => {
  if (!enabled) {
    couponCode.value = ''
  }
})

watch([selectedProductKey, quantity, selectedChannelId, couponEnabled, couponCode], () => {
  void loadPriceInfo()
})

onMounted(async () => {
  if (!appStore.publicSettingsLoaded) {
    try {
      await appStore.fetchPublicSettings()
    } catch {
      // Keep the page usable even if public settings fail to refresh.
    }
  }

  if (typeof window !== 'undefined') {
    agreementVisible.value = window.sessionStorage.getItem(AGREEMENT_STORAGE_KEY) !== '1'
  }

  if (!contact.value) {
    contact.value = authStore.user?.email ?? ''
  }

  await Promise.all([initializeStore(), refreshShareQrCode()])
})
</script>

<style scoped>
.store-page {
  min-height: 100vh;
  background: #edf2f7;
  color: #545454;
  padding-bottom: 140px;
}

.store-page-embedded {
  padding-top: 0;
}

.container {
  width: min(1120px, calc(100% - 32px));
  margin: 0 auto;
}

.header {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  background-color: #fff;
  z-index: 99;
  transition: 0.5s;
  box-shadow: 0 1px 15px 0 rgba(0, 0, 0, 0.12);
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header_left,
.header_right {
  display: flex;
  align-items: center;
  min-height: 70px;
}

.header_logo {
  display: flex;
  align-items: center;
  border: none;
  background: transparent;
  padding: 0;
  margin-right: 30px;
  height: auto;
  font-weight: 700;
  color: #535761;
  font-size: 18px;
  cursor: pointer;
}

.header_logo img {
  width: 48px;
  height: 48px;
  margin-right: 25px;
  border-radius: 12px;
  object-fit: cover;
}

.online-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 85px;
  padding: 0 8px;
  height: 25px;
  background: #698df3;
  border: 2px solid #4274ff;
  box-shadow: 1px 4px 5px 0 hsla(0, 0%, 66.3%, 0.2);
  border-radius: 12px;
  font-size: 12px;
  color: #fff;
  text-align: center;
  cursor: pointer;
  margin-right: 30px;
}

.header_button {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 130px;
  height: 44px;
  border: none;
  border-radius: 50px;
  background: linear-gradient(0deg, #2a62ff, #4e7dff);
  box-shadow: 0 5px 6px 0 rgba(73, 105, 230, 0.22);
}

.header_button span {
  display: block;
  font-size: 14px;
  color: #fff;
  letter-spacing: 0.4px;
  margin-left: 5px;
}

.section {
  position: relative;
  padding-top: 60px;
}

.section--first {
  padding-top: 110px;
}

.section--last {
  padding-bottom: 50px;
}

.merchant-card,
.main-card {
  box-shadow: 0 7px 29px 0 rgba(18, 52, 91, 0.11);
  border-radius: 10px;
  padding: 20px 35px;
  background: #fff;
}

.merchant-card {
  margin-bottom: 16px;
}

.section__title-wrap {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: flex-start;
  margin-top: 20px;
}

.section__title {
  color: #545454;
  font-weight: 700;
  font-size: 18px;
  line-height: 100%;
  margin-bottom: 0;
  position: relative;
  padding-left: 15px;
  font-family: 'Montserrat', sans-serif;
}

.section__title2,
.section .section__title_2 {
  font-size: 14px;
  color: #545454;
}

.section__title::before {
  content: '';
  position: absolute;
  display: block;
  top: 2px;
  bottom: 2px;
  left: 0;
  width: 3px;
  background-color: #3369ff;
  border-radius: 4px;
}

.section__title--pre::before {
  background-color: #f26c2a;
}

.section hr {
  border: none;
  height: 1px;
  background-color: #f7f7f7;
}

.merchant-loading {
  min-height: 120px;
}

.merchant-meta-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
}

.merchant-meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #545454;
  font-size: 14px;
  font-weight: 500;
}

.merchant-meta-item strong {
  color: #3369ff;
  font-weight: 600;
}

.merchant-meta-svg {
  width: 16px;
  height: 16px;
  flex: none;
}

.search {
  display: flex;
  flex-direction: row;
  align-items: center;
}

.search .input,
#order_box .ure_info_box .ure_info .ure_info_input_box .input {
  background: #fff;
  border: 1px solid #f0f0f0;
  box-shadow: 0 4px 10px 0 rgba(135, 142, 154, 0.07);
  border-radius: 4px;
  overflow: hidden;
}

.search .input {
  height: 40px;
  width: 400px;
  margin-right: 16px;
}

.search .input input,
#order_box .ure_info_box .ure_info .ure_info_input_box .input input {
  display: inline-block;
  width: 100%;
  padding: 0 20px;
  height: 100%;
  font-weight: 500;
  border: none;
  outline: none;
}

.search_button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100px;
  height: 32px;
  border: none;
  border-radius: 16px;
  background: linear-gradient(0deg, #2a62ff, #4e7dff);
  box-shadow: 0 5px 6px 0 rgba(73, 105, 230, 0.22);
  color: #ffffff;
  cursor: pointer;
  font-size: 14px;
}

.category_list {
  display: flex;
  gap: 16px;
  max-height: 420px;
  overflow-x: auto;
  padding-bottom: 6px;
}

.category_box,
.goods_box {
  padding: 0;
  border: none;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.category_box {
  min-width: 220px;
}

.goods_list {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  max-height: 420px;
  overflow-x: auto;
}

.card {
  position: relative;
  display: block;
  margin-top: 20px;
  border-radius: 10px;
  overflow: hidden;
  background-color: #fff;
  transition: 0.3s;
  border: 2px solid #f1f4fb;
  box-shadow: 0 4px 10px 0 rgba(135, 142, 154, 0.14);
}

.card:hover {
  box-shadow: 0 1px 15px 0 rgba(0, 0, 0, 0.2);
}

.card__title h3 {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  width: 100%;
  margin-bottom: 10px;
  color: #33373f;
  font-size: 15px;
  font-weight: 600;
}

.card__content,
.card__detail {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.card__s_cateremark {
  color: #999;
  font-size: 15px;
}

.category_box .card {
  min-height: 124px;
  padding: 1rem;
}

.category_box.active .card {
  background: linear-gradient(-45deg, #3369ff, #3798f7);
  box-shadow: 0 7px 10px 0 rgba(54, 144, 248, 0.23);
  border: 1px solid #3580fb;
}

.category_box.active .card .card__title h3,
.category_box.active .card .card__content .card__s_cateremark {
  color: #fff;
}

.goods_box .card {
  min-height: 124px;
  padding: 1rem;
}

.goods_box.active .card {
  border: 2px solid #3369ff;
}

.card__detail_price {
  font-weight: 600;
  font-size: 18px;
  color: #3369ff;
  line-height: 1.5;
}

.card__detail_stock {
  color: #0db26a;
  font-size: 14px;
}

.category-lite_img {
  position: absolute;
  right: -6px;
  bottom: -19px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
}

.category-lite_img::before {
  content: '';
  position: absolute;
  inset: 12px;
  border-right: 3px solid #dbe8ff;
  border-bottom: 3px solid #dbe8ff;
  transform: rotate(40deg);
}

.goods-lite_img,
.xiadui {
  position: absolute;
  right: -2px;
  bottom: -1px;
  width: 22px;
  height: 18px;
  border-radius: 0 0 4px 0;
  background: linear-gradient(135deg, #77a0ff, #3369ff);
}

.goods-lite_img::before,
.xiadui::before {
  content: '';
  position: absolute;
  right: 5px;
  bottom: 4px;
  width: 6px;
  height: 10px;
  border-right: 2px solid #fff;
  border-bottom: 2px solid #fff;
  transform: rotate(40deg);
}

.title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.text_box {
  padding: 20px;
  background: #fff;
  border: 1px solid #f0f4fb;
  border-radius: 10px;
  color: #666;
  line-height: 24px;
  font-size: 14px;
}

.text_box p {
  margin: 0;
  white-space: pre-wrap;
}

.sale_message span {
  display: block;
  font-size: 14px;
}

.sale_message span + span {
  margin-top: 8px;
}

#order_box .ure_info_box .ure_info_hide span {
  margin-left: 3px;
  color: #545454;
  font-size: 16px;
  font-weight: 700;
}

#order_box .ure_info_box .ure_info {
  margin: 0 auto;
  padding-top: 27px;
  padding-bottom: 25px;
  border-bottom: 1px solid #f7f7f7;
}

#order_box .ure_info_box .ure_info .ure_info_input_box {
  margin-left: 7px;
  margin-top: 20px;
}

#order_box .ure_info_box .ure_info .ure_info_input_box:first-child {
  margin-top: 0;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .ure_info_input_box_hide {
  font-weight: 700;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .ure_info_input_box_hide p {
  display: inline-block;
  vertical-align: middle;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .ure_info_input_box_hide p:first-child {
  font-size: 12px;
  color: #fb636b;
  width: 20px;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .ure_info_input_box_hide p:nth-child(2) {
  color: #545454;
  font-size: 14px;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .input {
  height: 50px;
  width: 430px;
  margin-top: 11px;
}

#order_box .ure_info_box .ure_info .ure_info_input_box .msg {
  margin-top: 14px;
  margin-left: 6px;
  font-size: 12px;
  font-weight: 500;
  color: #999;
}

.subscription-entry-box {
  padding: 27px 0 25px;
}

.subscription-entry-content {
  padding: 0 22px;
  color: #666;
  line-height: 1.9;
  font-size: 14px;
}

.subscription-entry-content p {
  margin: 0;
}

.subscription-entry-content p + p {
  margin-top: 10px;
}

.subscription-entry-title {
  font-size: 18px;
  font-weight: 700;
  color: #2f3a56;
}

.subscription-entry-hint {
  color: #d97706;
}

.purchase-toggle-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.btn-type {
  display: flex;
  align-items: center;
}

.btn-type > div {
  margin-right: 10px;
  border-radius: 45px;
  background: #f8f8f8;
  border: 2px solid #f6f6f6;
  color: #999;
  cursor: pointer;
}

.btn-type > div.on {
  border-color: #3369ff;
  color: #3369ff;
}

.btn-type > div.disabled {
  opacity: 0.55;
}

.btn-type label {
  display: inline-block;
  padding: 8px 16px;
  cursor: pointer;
  margin-bottom: 0;
  font-size: 14px;
}

#order_box .ure_info_box .pay_type .pay_type_hide {
  margin-left: 22px;
  padding-top: 22px;
  color: #999;
  font-size: 14px;
  font-weight: 700;
}

#order_box .ure_info_box .pay_type .pay_type_box {
  margin-top: 21px;
  margin-left: 17px;
  display: flex;
  flex-wrap: wrap;
  gap: 11px;
}

#order_box .ure_info_box .pay_type .pay_type_box .pay_type_leng {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 43px;
  width: 130px;
  min-height: 47px;
  text-align: center;
  background: #f8f8f8;
  border: 2px solid #f6f6f6;
  border-radius: 4px;
  position: relative;
  margin-bottom: 8px;
  cursor: pointer;
}

#order_box .ure_info_box .pay_type .pay_type_box .pay_type_leng img {
  width: 21px;
  height: 21px;
}

#order_box .ure_info_box .pay_type .pay_type_box .pay_type_leng span {
  font-weight: 700;
  font-size: 13px;
  color: #545454;
  margin-left: 8px;
}

#order_box .ure_info_box .pay_type .pay_type_box .pay_type_leng_xz {
  border-color: #3369ff;
}

#order_box .ure_info_box .pay_type .pay_type_box .pay_type_leng_xz span {
  color: #3369ff;
}

.section_buttom {
  width: 100%;
  height: 62px;
  background: #fff;
  box-shadow: 0 -1px 0 0 hsla(0, 0%, 71%, 0.18);
  position: fixed;
  left: 0;
  bottom: 0;
  z-index: 80;
}

.bottom-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 62px;
}

.goods_name {
  display: inline-flex;
  align-items: center;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.goods_name span {
  color: #545454;
  font-weight: 700;
  font-size: 14px;
  margin-left: 20px;
  max-width: 220px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.shuliang_box {
  display: inline-flex;
  align-items: center;
}

.subscription-bottom-summary {
  display: inline-flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  min-height: 40px;
  padding: 0 2px;
  color: #3369ff;
  font-size: 14px;
  font-weight: 600;
}

.shuliang_box .btn {
  width: 30px;
  height: 30px;
  background: #f8f8f8;
  border: 1px solid #f6f6f6;
  border-radius: 50%;
  position: relative;
  cursor: pointer;
  user-select: none;
}

.shuliang_box .input {
  margin-left: 20px;
  margin-right: 20px;
  background: #f8f8f8;
  border: 1px solid #f0f0f0;
  border-radius: 4px;
  height: 37px;
  width: 80px;
}

.shuliang_box .input input {
  background: #f8f8f8;
  border: none;
  width: 100%;
  color: #545454;
  font-weight: 700;
  font-size: 18px;
  height: 100%;
  text-align: center;
  outline: none;
}

.shuliang_box .input input::-webkit-outer-spin-button,
.shuliang_box .input input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.shuliang_box .btn span:first-child {
  width: 14px;
  height: 2px;
  margin-left: -7px;
  margin-top: -1px;
}

.shuliang_box .btn span:nth-child(2) {
  width: 2px;
  height: 14px;
  margin-left: -1px;
  margin-top: -7px;
}

.shuliang_box .btn span {
  background: #545454;
  position: absolute;
  left: 50%;
  top: 50%;
}

.minus-bar {
  display: block;
}

.jiage {
  margin-left: 50px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.jiage span:first-child {
  font-size: 12px;
  color: #545454;
  font-weight: 700;
}

.jiage span:nth-child(2),
.jiage > span:nth-child(4) {
  font-size: 22px;
  color: #3369ff;
  font-weight: 700;
}

.jiage s {
  font-size: 12px;
  color: #b7b7b7;
  margin-left: 12px;
}

.payment-fee-text {
  font-size: 12px;
  color: #666;
  margin-left: 2px;
}

.queding_btn {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 130px;
  height: 40px;
  border-radius: 50px;
  background: #3369ff;
  box-shadow: 0 5px 6px 0 rgba(73, 105, 230, 0.22);
  color: #fff;
}

.check_pay {
  width: 100%;
  height: 100%;
  border: none;
  background: transparent;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.check_pay:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.ewm {
  position: fixed;
  right: 2px;
  bottom: 50%;
  padding: 15px;
  background: #fff;
  border: 1px solid #eee;
  color: #666;
  line-height: 20px;
  font-size: 14px;
  text-align: center;
  z-index: 20;
}

.ewm img {
  width: 150px;
  height: 150px;
  display: block;
  margin-bottom: 8px;
}

footer {
  margin-bottom: 85px;
}

footer p {
  text-align: center;
  font-size: 14px;
  color: #585963;
}

.agreement-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.38);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 120;
  padding: 20px;
}

.agreement-modal {
  width: min(360px, 100%);
  max-height: min(540px, 85vh);
  background: #fff;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 24px 80px rgba(15, 23, 42, 0.24);
  display: flex;
  flex-direction: column;
}

.agreement-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
}

.agreement-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #4c5b78;
}

.agreement-close {
  border: none;
  background: transparent;
  font-size: 26px;
  line-height: 1;
  color: #475569;
  cursor: pointer;
}

.agreement-body {
  padding: 18px 20px;
  overflow: auto;
  font-size: 14px;
  line-height: 2;
  color: #333;
}

.agreement-body p {
  margin: 0 0 14px;
}

.agreement-emphasis {
  color: #ef4444;
  font-weight: 700;
}

.agreement-body ol {
  margin: 0;
  padding-left: 22px;
}

.agreement-body li + li {
  margin-top: 10px;
}

.agreement-footer {
  padding: 14px 20px 20px;
  display: flex;
  justify-content: flex-end;
}

.agreement-footer button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 110px;
  height: 36px;
  border: none;
  border-radius: 4px;
  background: linear-gradient(180deg, #4e7dff, #2a62ff);
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.store-loading,
.panel-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 180px;
  color: #64748b;
  font-size: 15px;
  font-weight: 600;
}

.loading-dot {
  width: 14px;
  height: 14px;
  border-radius: 999px;
  background: #3369ff;
  margin-right: 10px;
  box-shadow: 0 0 0 8px rgba(51, 105, 255, 0.12);
  animation: pulse 1.2s ease-in-out infinite;
}

.bangz_box {
  position: fixed;
  left: 0;
  top: 100px;
  z-index: 45;
}

.bangz_box .item {
  display: block;
  padding: 10px;
  border: none;
  border-radius: 0 7px 7px 0;
  margin-top: 25px;
  color: #fff;
  text-decoration: none;
  box-shadow: 0 7px 10px 0 rgba(54, 144, 248, 0.23);
}

.bangz_box .action-button {
  width: auto;
  font: inherit;
}

.bangz_box .item-danger {
  background: linear-gradient(45deg, #fd0b27, #ff4a4a);
  box-shadow: 0 7px 10px 0 rgba(255, 113, 19, 0.23);
}

.bangz_box .item-support {
  background: linear-gradient(45deg, #f737e8, #3369ff);
}

.bangz_box .item-group {
  background: linear-gradient(45deg, #3798f7, #3369ff);
}

.bangz_box span {
  margin-left: 5px;
  font-weight: 600;
  font-size: 14px;
  color: #fff;
}

@keyframes pulse {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.08);
  }
}

@media (max-width: 1100px) {
  .goods_list {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .bangz_box,
  .ewm {
    display: none;
  }
}

@media (max-width: 900px) {
  .header-inner {
    flex-wrap: wrap;
    justify-content: center;
    padding: 10px 0;
  }

  .header_left,
  .header_right {
    min-height: auto;
  }

  .section--first {
    padding-top: 145px;
  }

  .merchant-card,
  .main-card {
    padding: 18px 18px 24px;
  }

  .search {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .search .input,
  #order_box .ure_info_box .ure_info .ure_info_input_box .input {
    width: 100%;
    margin-right: 0;
  }

  .search_button {
    width: 100%;
    height: 40px;
    border-radius: 12px;
  }

  .goods_list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bottom-inner {
    flex-wrap: wrap;
    padding: 12px 0;
    gap: 12px;
  }

  .goods_name,
  .shuliang_box,
  .queding_btn {
    width: 100%;
  }

  .shuliang_box {
    justify-content: space-between;
    flex-wrap: wrap;
    row-gap: 12px;
  }

  .jiage {
    margin-left: 0;
    width: 100%;
    justify-content: flex-start;
  }
}

@media (max-width: 640px) {
  .container {
    width: min(100%, calc(100% - 20px));
  }

  .header_logo {
    margin-right: 0;
    font-size: 16px;
  }

  .header_logo img {
    width: 40px;
    height: 40px;
    margin-right: 14px;
  }

  .online-btn {
    width: 100%;
    justify-content: center;
    margin-top: 8px;
    margin-right: 0;
  }

  .section--first {
    padding-top: 170px;
  }

  .category_box {
    min-width: 180px;
  }

  .goods_list {
    grid-template-columns: 1fr;
  }

  .merchant-meta-row {
    gap: 10px;
  }

  .merchant-meta-item {
    width: 100%;
  }

  .goods_name span {
    max-width: calc(100vw - 90px);
  }
}
</style>
