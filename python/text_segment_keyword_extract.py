import nltk
import jieba
from keybert import KeyBERT
import json
import re
import sys
from typing import List, Dict

# 初始化模型
kw_model = KeyBERT(model="paraphrase-multilingual-MiniLM-L12-v2")  # 支持中文
nltk.download('punkt', quiet=True)

def segment_text(text: str) -> List[str]:
    """将长文本分段（按章节或段落）"""
    # 按章节分（匹配“第一章”“第1章”等）
    chapters = re.split(r'\n{2,}|\s*第[一二三四五六七八九十百]+章\s*|\s*第\d+章\s*', text)
    chapters = [ch.strip() for ch in chapters if ch.strip()]
    
    # 每章按段落或句子分
    segments = []
    for chapter in chapters:
        paragraphs = chapter.split('\n')
        for para in paragraphs:
            if para.strip():
                sentences = nltk.sent_tokenize(para)
                segments.extend(sentences)
    return [seg for seg in segments if len(seg) > 10]  # 过滤短句

def extract_keywords(text: str, top_n: int = 5) -> List[Dict]:
    """提取关键词"""
    words = jieba.lcut(text)
    processed_text = " ".join(words)
    keywords = kw_model.extract_keywords(
        processed_text,
        keyphrase_ngram_range=(1, 3),
        stop_words=None,
        top_n=top_n
    )
    return [{"keyword": kw[0], "score": kw[1]} for kw in keywords]

def process_novel(input_file: str, output_file: str = "output.json") -> Dict:
    """处理小说文本：分段并提取关键词"""
    with open(input_file, 'r', encoding='utf-8') as f:
        text = f.read()
    
    segments = segment_text(text)
    result = []
    for i, segment in enumerate(segments):
        keywords = extract_keywords(segment)
        result.append({
            "segment_id": i + 1,
            "text": segment,
            "keywords": keywords
        })
    
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(result, f, ensure_ascii=False, indent=2)
    
    return result

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("请提供输入文件路径")
        sys.exit(1)
    input_file = sys.argv[1]
    result = process_novel(input_file)

    # 关键：重新配置标准输出为 UTF-8
    sys.stdout.reconfigure(encoding='utf-8')

    print(json.dumps(result, ensure_ascii=False, indent=2))
