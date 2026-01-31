"""
LLM Answer Quality Evaluation System
Analyzes semantic similarity between LLM answers and expected answers
using multiple metrics and creates visualizations with Seaborn.
"""

import json
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
from typing import Dict, List, Tuple
from difflib import SequenceMatcher
import re
from collections import Counter

# Set style for better-looking plots
sns.set_style("whitegrid")
sns.set_palette("husl")
plt.rcParams['figure.figsize'] = (14, 10)


class LLMEvaluator:
    """Evaluates LLM answers against expected answers using multiple metrics."""
    
    def __init__(self, json_file_path: str):
        """Initialize evaluator with test results JSON file."""
        with open(json_file_path, 'r') as f:
            self.data = json.load(f)
        self.results = self.data['results']
        self.evaluation_df = None
        
    def extract_keywords(self, text: str) -> set:
        """Extract meaningful keywords from text."""
        # Remove common words and extract important terms
        stop_words = {'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 
                      'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were', 'be',
                      'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will',
                      'would', 'should', 'could', 'may', 'might', 'must', 'can', 'all'}
        
        # Convert to lowercase and split
        words = re.findall(r'\b\w+\b', text.lower())
        return set(word for word in words if word not in stop_words and len(word) > 2)
    
    def calculate_keyword_overlap(self, expected: str, actual: str) -> float:
        """Calculate keyword overlap score (Jaccard similarity)."""
        expected_keywords = self.extract_keywords(expected)
        actual_keywords = self.extract_keywords(actual)
        
        if not expected_keywords:
            return 0.0
            
        intersection = expected_keywords & actual_keywords
        union = expected_keywords | actual_keywords
        
        return len(intersection) / len(union) if union else 0.0
    
    def calculate_sequence_similarity(self, expected: str, actual: str) -> float:
        """Calculate sequence similarity using SequenceMatcher."""
        return SequenceMatcher(None, expected.lower(), actual.lower()).ratio()
    
    def calculate_length_ratio(self, expected: str, actual: str) -> float:
        """Calculate how close the lengths are (1.0 = same length)."""
        exp_len = len(expected)
        act_len = len(actual)
        
        if exp_len == 0:
            return 0.0
            
        ratio = min(exp_len, act_len) / max(exp_len, act_len)
        return ratio
    
    def calculate_number_accuracy(self, expected: str, actual: str) -> float:
        """Check if numerical values match."""
        exp_numbers = set(re.findall(r'\d+', expected))
        act_numbers = set(re.findall(r'\d+', actual))
        
        if not exp_numbers:
            return 1.0  # No numbers to match
            
        matches = exp_numbers & act_numbers
        return len(matches) / len(exp_numbers)
    
    def calculate_technical_term_match(self, expected: str, actual: str) -> float:
        """Match technical terms (error names, HTTP codes, etc.)."""
        # Technical patterns
        patterns = [
            r'\b\d{3}\b',  # HTTP status codes
            r'\b[A-Z][a-zA-Z]*Error\b',  # Error class names
            r'\b[A-Z][a-zA-Z]*Exception\b',  # Exception names
            r'\b(?:GET|POST|PUT|DELETE|PATCH)\b',  # HTTP methods
            r'/api/[\w/-]+',  # API endpoints
        ]
        
        exp_terms = set()
        act_terms = set()
        
        for pattern in patterns:
            exp_terms.update(re.findall(pattern, expected))
            act_terms.update(re.findall(pattern, actual))
        
        if not exp_terms:
            return 1.0
            
        matches = exp_terms & act_terms
        return len(matches) / len(exp_terms)
    
    def evaluate_answer_component(self, expected: str, actual: str) -> Dict[str, float]:
        """Evaluate a single component (root_cause, impact, or action) with multiple metrics."""
        return {
            'keyword_overlap': self.calculate_keyword_overlap(expected, actual),
            'sequence_similarity': self.calculate_sequence_similarity(expected, actual),
            'length_ratio': self.calculate_length_ratio(expected, actual),
            'number_accuracy': self.calculate_number_accuracy(expected, actual),
            'technical_term_match': self.calculate_technical_term_match(expected, actual)
        }
    
    def calculate_composite_score(self, metrics: Dict[str, float]) -> float:
        """Calculate weighted composite score from individual metrics."""
        weights = {
            'keyword_overlap': 0.30,
            'sequence_similarity': 0.20,
            'length_ratio': 0.10,
            'number_accuracy': 0.20,
            'technical_term_match': 0.20
        }
        
        return sum(metrics[key] * weights[key] for key in weights)
    
    def evaluate_all_tests(self) -> pd.DataFrame:
        """Evaluate all test results and return DataFrame."""
        evaluation_data = []
        
        for result in self.results:
            test_name = result['test_name']
            expected = result['expected_answer']
            actual = result['llm_answer']
            
            # Evaluate each component
            root_cause_metrics = self.evaluate_answer_component(
                expected['root_cause'], actual['root_cause']
            )
            impact_metrics = self.evaluate_answer_component(
                expected['impact'], actual['impact']
            )
            action_metrics = self.evaluate_answer_component(
                expected['suggested_action'], actual['suggested_action']
            )
            
            # Calculate composite scores
            root_cause_score = self.calculate_composite_score(root_cause_metrics)
            impact_score = self.calculate_composite_score(impact_metrics)
            action_score = self.calculate_composite_score(action_metrics)
            overall_score = (root_cause_score + impact_score + action_score) / 3
            
            evaluation_data.append({
                'test_name': test_name,
                'root_cause_score': root_cause_score,
                'impact_score': impact_score,
                'action_score': action_score,
                'overall_score': overall_score,
                'response_time': result['response_time_seconds'],
                
                # Individual metrics for root cause
                'rc_keyword_overlap': root_cause_metrics['keyword_overlap'],
                'rc_sequence_sim': root_cause_metrics['sequence_similarity'],
                'rc_technical_match': root_cause_metrics['technical_term_match'],
                
                # Individual metrics for impact
                'imp_keyword_overlap': impact_metrics['keyword_overlap'],
                'imp_sequence_sim': impact_metrics['sequence_similarity'],
                'imp_technical_match': impact_metrics['technical_term_match'],
                
                # Individual metrics for action
                'act_keyword_overlap': action_metrics['keyword_overlap'],
                'act_sequence_sim': action_metrics['sequence_similarity'],
                'act_technical_match': action_metrics['technical_term_match'],
            })
        
        self.evaluation_df = pd.DataFrame(evaluation_data)
        return self.evaluation_df
    
    def create_visualizations(self, output_dir: str = './metrics-assements'):
        """Create comprehensive visualizations of the evaluation results."""
        # Create output directory if it doesn't exist
        import os
        os.makedirs(output_dir, exist_ok=True)
        
        if self.evaluation_df is None:
            self.evaluate_all_tests()
        
        # Create figure with subplots
        fig = plt.figure(figsize=(18, 12))
        
        # 1. Overall Scores by Component (Grouped Bar Chart)
        ax1 = plt.subplot(2, 3, 1)
        score_data = self.evaluation_df[['root_cause_score', 'impact_score', 'action_score']].mean()
        colors = sns.color_palette("husl", 3)
        bars = ax1.bar(range(len(score_data)), score_data, color=colors, alpha=0.8, edgecolor='black')
        ax1.set_xticks(range(len(score_data)))
        ax1.set_xticklabels(['Root Cause', 'Impact', 'Action'], rotation=0)
        ax1.set_ylabel('Average Score', fontsize=11, fontweight='bold')
        ax1.set_title('Average Scores by Component', fontsize=12, fontweight='bold')
        ax1.set_ylim(0, 1)
        ax1.axhline(y=0.8, color='green', linestyle='--', alpha=0.5, label='Good (0.8)')
        ax1.axhline(y=0.6, color='orange', linestyle='--', alpha=0.5, label='Acceptable (0.6)')
        ax1.legend(fontsize=8)
        
        # Add value labels on bars
        for bar in bars:
            height = bar.get_height()
            ax1.text(bar.get_x() + bar.get_width()/2., height,
                    f'{height:.3f}', ha='center', va='bottom', fontsize=9, fontweight='bold')
        
        # 2. Score Distribution Heatmap
        ax2 = plt.subplot(2, 3, 2)
        score_matrix = self.evaluation_df[['root_cause_score', 'impact_score', 'action_score']].T
        score_matrix.columns = [f"T{i+1}" for i in range(len(self.evaluation_df))]
        sns.heatmap(score_matrix, annot=True, fmt='.2f', cmap='RdYlGn', 
                    vmin=0, vmax=1, cbar_kws={'label': 'Score'}, ax=ax2,
                    linewidths=0.5, linecolor='gray')
        ax2.set_ylabel('Component', fontsize=11, fontweight='bold')
        ax2.set_xlabel('Test Number', fontsize=11, fontweight='bold')
        ax2.set_title('Score Heatmap Across Tests', fontsize=12, fontweight='bold')
        ax2.set_yticklabels(['Root Cause', 'Impact', 'Action'], rotation=0)
        
        # 3. Overall Score per Test
        ax3 = plt.subplot(2, 3, 3)
        test_names_short = [f"T{i+1}" for i in range(len(self.evaluation_df))]
        colors_gradient = sns.color_palette("coolwarm", len(self.evaluation_df))
        bars = ax3.barh(test_names_short, self.evaluation_df['overall_score'], 
                        color=colors_gradient, edgecolor='black', alpha=0.8)
        ax3.set_xlabel('Overall Score', fontsize=11, fontweight='bold')
        ax3.set_ylabel('Test', fontsize=11, fontweight='bold')
        ax3.set_title('Overall Score by Test', fontsize=12, fontweight='bold')
        ax3.set_xlim(0, 1)
        ax3.axvline(x=0.8, color='green', linestyle='--', alpha=0.5)
        ax3.axvline(x=0.6, color='orange', linestyle='--', alpha=0.5)
        
        # Add value labels
        for i, bar in enumerate(bars):
            width = bar.get_width()
            ax3.text(width, bar.get_y() + bar.get_height()/2.,
                    f'{width:.3f}', ha='left', va='center', fontsize=8, fontweight='bold')
        
        # 4. Metric Breakdown (Radar-style comparison)
        ax4 = plt.subplot(2, 3, 4)
        metric_map = {
            'keyword_overlap': 'keyword_overlap',
            'sequence_sim': 'sequence_sim',
            'technical_match': 'technical_match'
        }
        metric_labels = ['Keyword\nOverlap', 'Sequence\nSimilarity', 'Technical\nMatch']
        
        rc_metrics = [self.evaluation_df[f'rc_{m}'].mean() for m in metric_map.values()]
        imp_metrics = [self.evaluation_df[f'imp_{m}'].mean() for m in metric_map.values()]
        act_metrics = [self.evaluation_df[f'act_{m}'].mean() for m in metric_map.values()]
        
        x = np.arange(len(metric_labels))
        width = 0.25
        
        ax4.bar(x - width, rc_metrics, width, label='Root Cause', alpha=0.8, edgecolor='black')
        ax4.bar(x, imp_metrics, width, label='Impact', alpha=0.8, edgecolor='black')
        ax4.bar(x + width, act_metrics, width, label='Action', alpha=0.8, edgecolor='black')
        
        ax4.set_ylabel('Average Score', fontsize=11, fontweight='bold')
        ax4.set_title('Metric Breakdown by Component', fontsize=12, fontweight='bold')
        ax4.set_xticks(x)
        ax4.set_xticklabels(metric_labels, fontsize=9)
        ax4.legend(fontsize=9)
        ax4.set_ylim(0, 1)
        ax4.grid(axis='y', alpha=0.3)
        
        # 5. Response Time vs Score
        ax5 = plt.subplot(2, 3, 5)
        scatter = ax5.scatter(self.evaluation_df['response_time'], 
                             self.evaluation_df['overall_score'],
                             c=self.evaluation_df['overall_score'], 
                             cmap='RdYlGn', s=150, alpha=0.7,
                             edgecolors='black', linewidth=1.5)
        ax5.set_xlabel('Response Time (seconds)', fontsize=11, fontweight='bold')
        ax5.set_ylabel('Overall Score', fontsize=11, fontweight='bold')
        ax5.set_title('Response Time vs Quality', fontsize=12, fontweight='bold')
        ax5.grid(alpha=0.3)
        
        # Add correlation coefficient
        corr = self.evaluation_df['response_time'].corr(self.evaluation_df['overall_score'])
        ax5.text(0.05, 0.95, f'Correlation: {corr:.3f}', 
                transform=ax5.transAxes, fontsize=10,
                verticalalignment='top', bbox=dict(boxstyle='round', facecolor='wheat', alpha=0.5))
        
        plt.colorbar(scatter, ax=ax5, label='Overall Score')
        
        # 6. Summary Statistics Table
        ax6 = plt.subplot(2, 3, 6)
        ax6.axis('tight')
        ax6.axis('off')
        
        summary_stats = pd.DataFrame({
            'Metric': ['Root Cause', 'Impact', 'Action', 'Overall'],
            'Mean': [
                self.evaluation_df['root_cause_score'].mean(),
                self.evaluation_df['impact_score'].mean(),
                self.evaluation_df['action_score'].mean(),
                self.evaluation_df['overall_score'].mean()
            ],
            'Std Dev': [
                self.evaluation_df['root_cause_score'].std(),
                self.evaluation_df['impact_score'].std(),
                self.evaluation_df['action_score'].std(),
                self.evaluation_df['overall_score'].std()
            ],
            'Min': [
                self.evaluation_df['root_cause_score'].min(),
                self.evaluation_df['impact_score'].min(),
                self.evaluation_df['action_score'].min(),
                self.evaluation_df['overall_score'].min()
            ],
            'Max': [
                self.evaluation_df['root_cause_score'].max(),
                self.evaluation_df['impact_score'].max(),
                self.evaluation_df['action_score'].max(),
                self.evaluation_df['overall_score'].max()
            ]
        })
        
        # Format values for display
        formatted_data = []
        for _, row in summary_stats.iterrows():
            formatted_data.append([
                row['Metric'],
                f"{row['Mean']:.3f}",
                f"{row['Std Dev']:.3f}",
                f"{row['Min']:.3f}",
                f"{row['Max']:.3f}"
            ])
        
        table = ax6.table(cellText=formatted_data, 
                         colLabels=summary_stats.columns,
                         cellLoc='center', loc='center',
                         colWidths=[0.3, 0.2, 0.2, 0.15, 0.15])
        table.auto_set_font_size(False)
        table.set_fontsize(9)
        table.scale(1, 2)
        
        # Style header
        for i in range(len(summary_stats.columns)):
            table[(0, i)].set_facecolor('#4CAF50')
            table[(0, i)].set_text_props(weight='bold', color='white')
        
        # Color code rows
        colors_table = ['#e8f5e9', '#fff9c4', '#ffe0b2', '#bbdefb']
        for i in range(1, len(summary_stats) + 1):
            for j in range(len(summary_stats.columns)):
                table[(i, j)].set_facecolor(colors_table[i-1])
        
        ax6.set_title('Summary Statistics', fontsize=12, fontweight='bold', pad=20)
        
        plt.tight_layout()
        plt.savefig(f'{output_dir}/llm_evaluation_dashboard.png', dpi=300, bbox_inches='tight')
        print(f"Dashboard saved to {output_dir}/llm_evaluation_dashboard.png")
        
        # Create detailed breakdown chart
        self._create_detailed_breakdown(output_dir)
        
        return fig
    
    def _create_detailed_breakdown(self, output_dir: str):
        """Create detailed breakdown of individual test performance."""
        fig, axes = plt.subplots(2, 5, figsize=(20, 8))
        axes = axes.flatten()
        
        for idx, (i, row) in enumerate(self.evaluation_df.iterrows()):
            ax = axes[idx]
            
            # Create spider/radar chart for each test
            categories = ['Root\nCause', 'Impact', 'Action']
            scores = [row['root_cause_score'], row['impact_score'], row['action_score']]
            
            # Create bar chart for each test
            colors = sns.color_palette("husl", 3)
            bars = ax.bar(categories, scores, color=colors, alpha=0.7, edgecolor='black')
            ax.set_ylim(0, 1)
            ax.set_title(f"Test {idx+1}\n{row['test_name'][:30]}...", fontsize=9, fontweight='bold')
            ax.axhline(y=0.8, color='green', linestyle='--', alpha=0.3)
            ax.axhline(y=0.6, color='orange', linestyle='--', alpha=0.3)
            ax.set_ylabel('Score', fontsize=8)
            ax.tick_params(axis='x', labelsize=7)
            ax.tick_params(axis='y', labelsize=7)
            
            # Add score values on bars
            for bar, score in zip(bars, scores):
                height = bar.get_height()
                ax.text(bar.get_x() + bar.get_width()/2., height,
                       f'{score:.2f}', ha='center', va='bottom', fontsize=7)
        
        plt.tight_layout()
        plt.savefig(f'{output_dir}/llm_evaluation_detailed.png', dpi=300, bbox_inches='tight')
        print(f"Detailed breakdown saved to {output_dir}/llm_evaluation_detailed.png")
    
    def generate_report(self, output_dir: str = './metrics-assements') -> str:
        """Generate a text report with insights."""
        # Create output directory if it doesn't exist
        import os
        os.makedirs(output_dir, exist_ok=True)
        
        if self.evaluation_df is None:
            self.evaluate_all_tests()
        
        report = []
        report.append("=" * 80)
        report.append("LLM ANSWER QUALITY EVALUATION REPORT")
        report.append("=" * 80)
        report.append("")
        
        # Overall Statistics
        report.append("OVERALL PERFORMANCE")
        report.append("-" * 80)
        report.append(f"Average Overall Score: {self.evaluation_df['overall_score'].mean():.3f}")
        report.append(f"Average Root Cause Score: {self.evaluation_df['root_cause_score'].mean():.3f}")
        report.append(f"Average Impact Score: {self.evaluation_df['impact_score'].mean():.3f}")
        report.append(f"Average Action Score: {self.evaluation_df['action_score'].mean():.3f}")
        report.append(f"Average Response Time: {self.evaluation_df['response_time'].mean():.2f} seconds")
        report.append("")
        
        # Performance Categories
        excellent = self.evaluation_df[self.evaluation_df['overall_score'] >= 0.85]
        good = self.evaluation_df[(self.evaluation_df['overall_score'] >= 0.70) & 
                                   (self.evaluation_df['overall_score'] < 0.85)]
        acceptable = self.evaluation_df[(self.evaluation_df['overall_score'] >= 0.60) & 
                                         (self.evaluation_df['overall_score'] < 0.70)]
        needs_improvement = self.evaluation_df[self.evaluation_df['overall_score'] < 0.60]
        
        report.append("PERFORMANCE DISTRIBUTION")
        report.append("-" * 80)
        report.append(f"Excellent (≥0.85): {len(excellent)} tests ({len(excellent)/len(self.evaluation_df)*100:.1f}%)")
        report.append(f"Good (0.70-0.84): {len(good)} tests ({len(good)/len(self.evaluation_df)*100:.1f}%)")
        report.append(f"Acceptable (0.60-0.69): {len(acceptable)} tests ({len(acceptable)/len(self.evaluation_df)*100:.1f}%)")
        report.append(f"Needs Improvement (<0.60): {len(needs_improvement)} tests ({len(needs_improvement)/len(self.evaluation_df)*100:.1f}%)")
        report.append("")
        
        # Best and Worst Performing Tests
        report.append("TOP 3 PERFORMING TESTS")
        report.append("-" * 80)
        top_tests = self.evaluation_df.nlargest(3, 'overall_score')
        for idx, row in top_tests.iterrows():
            report.append(f"{row['test_name']}: {row['overall_score']:.3f}")
            report.append(f"  - Root Cause: {row['root_cause_score']:.3f}, Impact: {row['impact_score']:.3f}, Action: {row['action_score']:.3f}")
        report.append("")
        
        report.append("BOTTOM 3 PERFORMING TESTS")
        report.append("-" * 80)
        bottom_tests = self.evaluation_df.nsmallest(3, 'overall_score')
        for idx, row in bottom_tests.iterrows():
            report.append(f"{row['test_name']}: {row['overall_score']:.3f}")
            report.append(f"  - Root Cause: {row['root_cause_score']:.3f}, Impact: {row['impact_score']:.3f}, Action: {row['action_score']:.3f}")
        report.append("")
        
        # Insights
        report.append("KEY INSIGHTS")
        report.append("-" * 80)
        
        # Which component performs best/worst
        comp_scores = {
            'Root Cause': self.evaluation_df['root_cause_score'].mean(),
            'Impact': self.evaluation_df['impact_score'].mean(),
            'Action': self.evaluation_df['action_score'].mean()
        }
        best_comp = max(comp_scores, key=comp_scores.get)
        worst_comp = min(comp_scores, key=comp_scores.get)
        
        report.append(f"• Strongest Component: {best_comp} ({comp_scores[best_comp]:.3f})")
        report.append(f"• Weakest Component: {worst_comp} ({comp_scores[worst_comp]:.3f})")
        
        # Response time correlation
        corr = self.evaluation_df['response_time'].corr(self.evaluation_df['overall_score'])
        if abs(corr) < 0.3:
            corr_interp = "weak"
        elif abs(corr) < 0.7:
            corr_interp = "moderate"
        else:
            corr_interp = "strong"
        
        report.append(f"• Response Time Correlation: {corr:.3f} ({corr_interp})")
        
        # Consistency
        std = self.evaluation_df['overall_score'].std()
        if std < 0.05:
            consistency = "very consistent"
        elif std < 0.10:
            consistency = "consistent"
        else:
            consistency = "variable"
        
        report.append(f"• Score Consistency: {consistency} (std: {std:.3f})")
        report.append("")
        
        report.append("=" * 80)
        
        report_text = "\n".join(report)
        
        # Save report
        with open(f'{output_dir}/evaluation_report.txt', 'w') as f:
            f.write(report_text)
        
        print(report_text)
        print(f"\nReport saved to {output_dir}/evaluation_report.txt")
        
        return report_text


def main():
    """Main execution function."""
    import os
    
    # Initialize evaluator
    evaluator = LLMEvaluator('./llm_test_results.json')
    
    # Run evaluation
    print("Evaluating LLM test results...")
    df = evaluator.evaluate_all_tests()
    
    # Display results
    print("\n" + "="*80)
    print("EVALUATION RESULTS SUMMARY")
    print("="*80)
    print(df[['test_name', 'root_cause_score', 'impact_score', 'action_score', 'overall_score']].to_string(index=False))
    
    # Create output directory
    output_dir = './metrics-assements'
    os.makedirs(output_dir, exist_ok=True)
    
    # Create visualizations
    print("\nGenerating visualizations...")
    evaluator.create_visualizations(output_dir)
    
    # Generate report
    print("\nGenerating detailed report...")
    evaluator.generate_report(output_dir)
    
    # Save detailed CSV
    df.to_csv(f'{output_dir}/detailed_evaluation_metrics.csv', index=False)
    print(f"\nDetailed metrics saved to {output_dir}/detailed_evaluation_metrics.csv")
    
    print("\n✅ Evaluation complete! Check the outputs directory for visualizations and reports.")


if __name__ == "__main__":
    main()