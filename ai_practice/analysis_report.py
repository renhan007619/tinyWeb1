#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
交通运输学院论文关键词分析报告
作者：重庆交通大学交通运输学院
功能：数据清洗、词频统计、词云生成、相似度计算
"""

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from wordcloud import WordCloud
import jieba
import jieba.analyse
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
import os
import warnings
warnings.filterwarnings('ignore')

# 设置中文字体
plt.rcParams['font.sans-serif'] = ['SimHei', 'Microsoft YaHei', 'Arial Unicode MS']
plt.rcParams['axes.unicode_minus'] = False

# 配置参数
DATA_PATH = 'data/论文信息.xlsx'
OUTPUT_DIR = 'output'
TARGET_COLLEGE = '交通运输学院'  # 目标学院
# 注：如需扩展到其他学院，修改上方 TARGET_COLLEGE 即可，无需改动下游统计逻辑

# 停用词列表
STOPWORDS = set([
    '的', '了', '在', '是', '我', '有', '和', '就', '不', '人', '都', '一', '一个', '上', '也',
    '很', '到', '说', '要', '去', '你', '会', '着', '没有', '看', '好', '自己', '这', '那',
    '这些', '那些', '这个', '那个', '之', '与', '及', '等', '或', '但', '而', '以', '为',
    '将', '于', '由', '从', '对', '关于', '基于', '研究', '分析', '设计', '系统', '方法',
    '技术', '应用', '实现', '优化', '探讨', '论', '试', '初探', '刍议', '浅析', '研究',
    '一个', '一种', '以及', '及其', '进行', '提出', '通过', '采用', '利用', '使用'
])


def ensure_output_dir():
    """确保输出目录存在"""
    if not os.path.exists(OUTPUT_DIR):
        os.makedirs(OUTPUT_DIR)
        print(f"✓ 创建输出目录: {OUTPUT_DIR}")


def load_and_clean_data(filepath):
    """
    1. 数据读取与清洗
    """
    print("\n" + "="*60)
    print("【第一步】数据读取与清洗")
    print("="*60)

    # 读取数据
    print(f"正在读取文件: {filepath}")
    df = pd.read_excel(filepath)
    print(f"✓ 成功读取数据，原始数据共 {len(df)} 行，{len(df.columns)} 列")

    # 显示列名
    print(f"\n数据列名: {list(df.columns)}")

    # 查看所有学院
    if '学院' in df.columns:
        colleges = df['学院'].unique()
        print(f"\n数据包含的学院: {colleges}")

    # 筛选目标学院数据
    if '学院' in df.columns:
        df_college = df[df['学院'] == TARGET_COLLEGE].copy()
        print(f"\n✓ 筛选出【{TARGET_COLLEGE}】的数据: {len(df_college)} 条")

        if len(df_college) == 0:
            print(f"⚠ 警告: 未找到{TARGET_COLLEGE}的数据，将分析所有学院数据")
            df_college = df.copy()
    else:
        print("⚠ 警告: 数据中没有'学院'列，将分析所有数据")
        df_college = df.copy()

    # 数据清洗
    print("\n数据清洗过程:")

    # 处理缺失值
    initial_rows = len(df_college)

    # 删除关键字段为空的行
    if '论文关键词' in df_college.columns:
        df_college = df_college.dropna(subset=['论文关键词'])
        print(f"  - 删除关键词为空的记录: {initial_rows - len(df_college)} 条")

    if '题目所属专业' in df_college.columns:
        df_college = df_college.dropna(subset=['题目所属专业'])
        print(f"  - 删除专业为空的记录后: {len(df_college)} 条")

    # 去除重复数据
    before_dedup = len(df_college)
    df_college = df_college.drop_duplicates()
    print(f"  - 去除重复记录: {before_dedup - len(df_college)} 条")

    # 清理文本数据
    text_columns = ['论文关键词', '论文题目', '论文研究方向']
    for col in text_columns:
        if col in df_college.columns:
            df_college[col] = df_college[col].astype(str).str.strip()
            # 去除多余的空格
            df_college[col] = df_college[col].str.replace(r'\s+', ' ', regex=True)

    print(f"\n✓ 清洗完成，最终数据: {len(df_college)} 条")

    return df_college


def visualize_data(df):
    """
    2. 数据可视化
    """
    print("\n" + "="*60)
    print("【第二步】数据可视化")
    print("="*60)

    if len(df) == 0:
        print("⚠ 数据为空，跳过可视化")
        return

    # 图1: 各专业论文数量统计
    if '题目所属专业' in df.columns:
        plt.figure(figsize=(12, 6))
        major_counts = df['题目所属专业'].value_counts()

        ax = sns.barplot(x=major_counts.index, y=major_counts.values, palette='viridis')
        plt.title(f'{TARGET_COLLEGE} - 各专业论文数量统计', fontsize=16, fontweight='bold')
        plt.xlabel('专业', fontsize=12)
        plt.ylabel('论文数量', fontsize=12)
        plt.xticks(rotation=45, ha='right')

        # 在柱状图上添加数值
        for i, v in enumerate(major_counts.values):
            ax.text(i, v + 0.5, str(v), ha='center', va='bottom', fontweight='bold')

        plt.tight_layout()
        plt.savefig(f'{OUTPUT_DIR}/专业论文数量统计.png', dpi=300, bbox_inches='tight')
        print(f"✓ 保存图表: {OUTPUT_DIR}/专业论文数量统计.png")
        plt.close()

        print(f"\n各专业论文数量:")
        for major, count in major_counts.items():
            print(f"  - {major}: {count} 篇")

    # 图2: 论文类型分布（饼图）
    if '论文类型' in df.columns:
        plt.figure(figsize=(10, 8))
        type_counts = df['论文类型'].value_counts()

        colors = plt.cm.Set3(np.linspace(0, 1, len(type_counts)))
        plt.pie(type_counts.values, labels=type_counts.index, autopct='%1.1f%%',
                colors=colors, startangle=90, textprops={'fontsize': 11})
        plt.title(f'{TARGET_COLLEGE} - 论文类型分布', fontsize=16, fontweight='bold')
        plt.axis('equal')

        plt.tight_layout()
        plt.savefig(f'{OUTPUT_DIR}/论文类型分布.png', dpi=300, bbox_inches='tight')
        print(f"✓ 保存图表: {OUTPUT_DIR}/论文类型分布.png")
        plt.close()


def word_frequency_analysis(df):
    """
    3. 词频统计
    """
    print("\n" + "="*60)
    print("【第三步】词频统计")
    print("="*60)

    if '论文关键词' not in df.columns:
        print("⚠ 数据中缺少'论文关键词'列，跳过词频统计")
        return {}

    print("正在提取关键词并统计词频...")

    # 合并所有关键词
    all_keywords = []
    major_keywords = {}  # 按专业分组的关键词

    for _, row in df.iterrows():
        major = row.get('题目所属专业', '未知专业')
        keywords_text = str(row.get('论文关键词', ''))

        # 分词处理
        # 关键词通常用分号、逗号、空格分隔
        keywords = re.split(r'[；;，,、\s]+', keywords_text)
        keywords = [k.strip() for k in keywords if k.strip() and len(k.strip()) > 1]

        all_keywords.extend(keywords)

        if major not in major_keywords:
            major_keywords[major] = []
        major_keywords[major].extend(keywords)

    # 整体词频统计
    from collections import Counter
    word_freq = Counter(all_keywords)

    print(f"\n✓ 共提取到 {len(all_keywords)} 个关键词")
    print(f"✓ 去重后共有 {len(word_freq)} 个不同的关键词")

    # 显示TOP20高频词
    print("\n【TOP 20 高频关键词】")
    for word, freq in word_freq.most_common(20):
        print(f"  {word}: {freq} 次")

    # 保存词频统计到Excel
    freq_df = pd.DataFrame(word_freq.most_common(), columns=['关键词', '频次'])
    freq_df.to_excel(f'{OUTPUT_DIR}/词频统计表.xlsx', index=False)
    print(f"\n✓ 词频统计表已保存: {OUTPUT_DIR}/词频统计表.xlsx")

    # 各专业词频统计
    print("\n【各专业TOP 5关键词】")
    for major, keywords in major_keywords.items():
        major_freq = Counter(keywords)
        print(f"\n{major}:")
        for word, freq in major_freq.most_common(5):
            print(f"  - {word}: {freq} 次")

    return word_freq


def generate_wordcloud(word_freq):
    """
    4. 生成词云图
    """
    print("\n" + "="*60)
    print("【第四步】生成词云图")
    print("="*60)

    if not word_freq:
        print("⚠ 词频数据为空，跳过词云生成")
        return

    print("正在生成词云图...")

    # 准备词云数据
    wordcloud_data = {word: freq for word, freq in word_freq.items() if freq >= 2}

    # 生成词云
    wordcloud = WordCloud(
        width=1200,
        height=800,
        background_color='white',
        font_path=None,  # 使用默认字体
        max_words=200,
        relative_scaling=0.5,
        colormap='viridis',
        prefer_horizontal=0.7,
        min_font_size=10,
        max_font_size=150
    ).generate_from_frequencies(wordcloud_data)

    # 保存词云图
    plt.figure(figsize=(15, 10))
    plt.imshow(wordcloud, interpolation='bilinear')
    plt.axis('off')
    plt.title(f'{TARGET_COLLEGE} - 论文关键词词云图', fontsize=20, fontweight='bold', pad=20)
    plt.tight_layout()
    plt.savefig(f'{OUTPUT_DIR}/词云图.png', dpi=300, bbox_inches='tight')
    print(f"✓ 词云图已保存: {OUTPUT_DIR}/词云图.png")
    plt.close()


def calculate_similarity(df):
    """
    5. 专业与关键词的相似度计算
    使用TF-IDF + 余弦相似度
    """
    print("\n" + "="*60)
    print("【第五步】专业与关键词相似度计算")
    print("="*60)

    if '论文关键词' not in df.columns or '题目所属专业' not in df.columns:
        print("⚠ 缺少必要列，跳过相似度计算")
        return

    print("使用TF-IDF和余弦相似度计算专业与关键词的相似度...")

    # 按专业聚合关键词
    major_texts = {}
    for major in df['题目所属专业'].unique():
        major_df = df[df['题目所属专业'] == major]
        keywords_list = []
        for keywords in major_df['论文关键词']:
            keywords_list.append(str(keywords))
        major_texts[major] = ' '.join(keywords_list)

    if len(major_texts) < 2:
        print("⚠ 专业数量不足，无法计算相似度")
        return

    # 准备文档
    majors = list(major_texts.keys())
    documents = list(major_texts.values())

    print(f"\n分析的专业: {majors}")

    # 使用TF-IDF向量化
    vectorizer = TfidfVectorizer(
        tokenizer=lambda x: jieba.lcut(x),
        stop_words=list(STOPWORDS),
        max_features=1000,
        min_df=1
    )

    try:
        tfidf_matrix = vectorizer.fit_transform(documents)
        feature_names = vectorizer.get_feature_names_out()

        print(f"\n✓ TF-IDF特征维度: {tfidf_matrix.shape}")

        # 计算余弦相似度
        similarity_matrix = cosine_similarity(tfidf_matrix)

        # 显示相似度矩阵
        print("\n【专业间相似度矩阵】")
        similarity_df = pd.DataFrame(
            similarity_matrix,
            index=majors,
            columns=majors
        )
        print(similarity_df.round(4))

        # 可视化相似度热力图
        plt.figure(figsize=(10, 8))
        sns.heatmap(similarity_df, annot=True, fmt='.3f', cmap='YlOrRd',
                    square=True, linewidths=0.5, cbar_kws={"shrink": .8})
        plt.title(f'{TARGET_COLLEGE} - 专业间关键词相似度热力图', fontsize=16, fontweight='bold')
        plt.tight_layout()
        plt.savefig(f'{OUTPUT_DIR}/专业相似度热力图.png', dpi=300, bbox_inches='tight')
        print(f"\n✓ 相似度热力图已保存: {OUTPUT_DIR}/专业相似度热力图.png")
        plt.close()

        # 分析每个专业的特征关键词
        print("\n【各专业特征关键词（TF-IDF权重TOP 10）】")
        for i, major in enumerate(majors):
            tfidf_scores = tfidf_matrix[i].toarray()[0]
            top_indices = tfidf_scores.argsort()[-10:][::-1]

            print(f"\n{major}:")
            for idx in top_indices:
                if tfidf_scores[idx] > 0:
                    print(f"  - {feature_names[idx]}: {tfidf_scores[idx]:.4f}")

    except Exception as e:
        print(f"⚠ 相似度计算出错: {e}")


def generate_report(df):
    """
    生成分析报告摘要
    """
    print("\n" + "="*60)
    print("【分析报告摘要】")
    print("="*60)

    print(f"""
一、数据概况
  - 分析对象: {TARGET_COLLEGE}
  - 论文总数: {len(df)} 篇
  - 专业数量: {df['题目所属专业'].nunique() if '题目所属专业' in df.columns else 'N/A'} 个

二、数据质量
  - 数据完整性: 已清洗缺失值和重复数据
  - 关键词提取: 完成分词和频率统计

三、主要发现
  1. 论文数量最多的专业: {df['题目所属专业'].value_counts().index[0] if '题目所属专业' in df.columns and len(df) > 0 else 'N/A'}
     共 {df['题目所属专业'].value_counts().iloc[0] if '题目所属专业' in df.columns and len(df) > 0 else 0} 篇

  2. 主要论文类型: {df['论文类型'].value_counts().index[0] if '论文类型' in df.columns and len(df) > 0 else 'N/A'}

四、输出文件
  - 专业论文数量统计.png
  - 论文类型分布.png
  - 词云图.png
  - 专业相似度热力图.png
  - 词频统计表.xlsx

五、相似度分析
  - 使用TF-IDF向量化方法
  - 采用余弦相似度计算专业间相关性
  - 结果已可视化保存
""")


# 导入re模块（放在文件顶部更好，但为了代码完整性放在这里）
import re


def main():
    """主函数"""
    print("\n" + "="*60)
    print("  重庆交通大学 - 交通运输学院论文关键词分析")
    print("="*60)

    # 确保输出目录存在
    ensure_output_dir()

    # 1. 数据读取与清洗
    df = load_and_clean_data(DATA_PATH)

    if len(df) == 0:
        print("\n⚠ 错误: 没有有效数据可供分析")
        return

    # 2. 数据可视化
    visualize_data(df)

    # 3. 词频统计
    word_freq = word_frequency_analysis(df)

    # 4. 生成词云图
    generate_wordcloud(word_freq)

    # 5. 相似度计算
    calculate_similarity(df)

    # 生成报告摘要
    generate_report(df)

    print("\n" + "="*60)
    print("  分析完成！所有结果已保存到 output/ 目录")
    print("="*60)


if __name__ == '__main__':
    main()
